package command

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Service объединяет обработчики команд приложения и управляет фоновыми задачами.
// Реализация каждой команды вынесена в отдельный файл пакета command.
type Service struct {
	writeRepo ports.ExaminationWriteRepository
	speech    ports.SpeechClient
	llm       ports.LLMClient
	storage   ports.ObjectStorage
	logger    *slog.Logger

	processingCtx    context.Context
	processingCancel context.CancelFunc
	queue            chan processingTask
	processing       errgroup.Group
	stateMu          sync.RWMutex
	closed           bool
}

// processingTask содержит неизменяемые данные задания, передаваемого worker'у.
type processingTask struct {
	examinationID uuid.UUID
	jobID         uuid.UUID
	objectKey     string
	fileName      string
	transcript    *string
	started       bool
}

// NewService создаёт фасад команд и запускает ограниченный пул worker'ов.
// workers задаёт число одновременно обрабатываемых заданий, queueSize —
// максимальное число заданий, ожидающих свободного worker'а.
func NewService(parent context.Context,
	writeRepo ports.ExaminationWriteRepository,
	speech ports.SpeechClient,
	llm ports.LLMClient,
	storage ports.ObjectStorage,
	logger *slog.Logger,
	workers int,
	queueSize int) *Service {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = workers
	}
	ctx, cancel := context.WithCancel(parent)
	service := &Service{
		writeRepo:        writeRepo,
		speech:           speech,
		llm:              llm,
		storage:          storage,
		logger:           logger,
		processingCtx:    ctx,
		processingCancel: cancel,
		queue:            make(chan processingTask, queueSize),
	}
	service.startWorkers(workers)
	service.startRecovery(queueSize)
	return service
}

// startWorkers запускает фиксированное число обработчиков очереди.
func (s *Service) startWorkers(workers int) {
	for range workers {
		s.processing.Go(func() error {
			for {
				select {
				case <-s.processingCtx.Done():
					return nil
				case task := <-s.queue:
					s.process(task)
				}
			}
		})
	}
}

// startRecovery запускает один координатор восстановления заданий из PostgreSQL.
func (s *Service) startRecovery(batchSize int) {
	s.processing.Go(func() error {
		// Фиксированная граница не позволяет повторно захватить задания,
		// которые уже были запущены этим проходом восстановления.
		staleBefore := time.Now().UTC()
		for {
			count, err := s.recoverPending(staleBefore, batchSize)
			if err != nil {
				if s.processingCtx.Err() == nil {
					s.logger.Error("не удалось восстановить фоновые задания", "error", err)
				}
				return nil
			}
			if count < batchSize {
				return nil
			}
		}
	})
}

func (s *Service) recoverPending(staleBefore time.Time, batchSize int) (int, error) {
	now := time.Now().UTC()
	tasks, err := s.writeRepo.ClaimPendingProcessing(s.processingCtx, staleBefore, now, batchSize)
	if err != nil {
		return 0, err
	}
	for _, recovered := range tasks {
		var transcript *string
		if strings.TrimSpace(recovered.Transcript) != "" {
			value := recovered.Transcript
			transcript = &value
		}
		task := processingTask{
			examinationID: recovered.ExaminationID,
			jobID:         recovered.JobID,
			objectKey:     recovered.ObjectKey,
			fileName:      recovered.FileName,
			transcript:    transcript,
			started:       true,
		}
		if transcript == nil && task.objectKey == "" {
			s.fail(task.examinationID, task.jobID, ports.ErrFileRequired)
			continue
		}
		if err := s.enqueueWait(task); err != nil {
			return 0, err
		}
		s.logger.Info("незавершённое задание восстановлено", "examination_id", task.examinationID, "job_id", task.jobID, "attempt", recovered.Attempt+1)
	}
	return len(tasks), nil
}

// enqueueWait ожидает место в очереди и используется только координатором
// восстановления, количество goroutine при этом остаётся фиксированным.
func (s *Service) enqueueWait(task processingTask) error {
	select {
	case <-s.processingCtx.Done():
		return s.processingCtx.Err()
	case s.queue <- task:
		return nil
	}
}

// enqueue добавляет задание в ограниченную очередь без создания новой goroutine.
func (s *Service) enqueue(task processingTask) error {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	if s.closed {
		return context.Canceled
	}
	select {
	case <-s.processingCtx.Done():
		return s.processingCtx.Err()
	case s.queue <- task:
		return nil
	default:
		return ports.ErrProcessingQueueFull
	}
}

// Close запрещает запуск новой фоновой работы, отменяет текущую и ожидает
// завершения всех зарегистрированных задач обработки.
func (s *Service) Close() error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()
		return nil
	}
	s.closed = true
	s.processingCancel()
	s.stateMu.Unlock()
	return s.processing.Wait()
}

var _ ports.ExaminationCommandHandler = (*Service)(nil)
