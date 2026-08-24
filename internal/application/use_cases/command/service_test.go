package command

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/google/uuid"
)

type repositoryStub struct {
	mu                sync.Mutex
	completed, failed chan struct{}
	pending           []ports.PendingProcessingTask
}

type blockingRepository struct {
	*repositoryStub
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

type concurrencyRepository struct {
	*repositoryStub
	active  atomic.Int32
	maximum atomic.Int32
	started chan struct{}
	release chan struct{}
}

func (r *concurrencyRepository) StartProcessing(ctx context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time, _ int) error {
	current := r.active.Add(1)
	defer r.active.Add(-1)
	for {
		maximum := r.maximum.Load()
		if current <= maximum || r.maximum.CompareAndSwap(maximum, current) {
			break
		}
	}
	r.started <- struct{}{}
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *blockingRepository) StartProcessing(ctx context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time, _ int) error {
	r.once.Do(func() { close(r.started) })
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *repositoryStub) CreateExamination(context.Context, ports.ExaminationWriteModel, ports.ProcessingJobWriteModel) error {
	return nil
}
func (r *repositoryStub) StartProcessing(context.Context, uuid.UUID, uuid.UUID, time.Time, int) error {
	return nil
}
func (r *repositoryStub) SaveTranscript(context.Context, uuid.UUID, uuid.UUID, string, time.Time) error {
	return nil
}
func (r *repositoryStub) SaveSummary(context.Context, uuid.UUID, string, time.Time) error { return nil }
func (r *repositoryStub) CompleteProcessing(context.Context, uuid.UUID, uuid.UUID, time.Time) error {
	r.completed <- struct{}{}
	return nil
}
func (r *repositoryStub) FailProcessing(context.Context, uuid.UUID, uuid.UUID, string, time.Time) error {
	r.failed <- struct{}{}
	return nil
}
func (r *repositoryStub) RetryProcessing(context.Context, int64, uuid.UUID, uuid.UUID, time.Time) (uuid.UUID, string, string, string, error) {
	return uuid.New(), "transcript", "", "", nil
}
func (r *repositoryStub) DeleteExamination(context.Context, int64, uuid.UUID) error {
	return nil
}
func (r *repositoryStub) ClaimPendingProcessing(context.Context, time.Time, time.Time, int) ([]ports.PendingProcessingTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tasks := r.pending
	r.pending = nil
	return tasks, nil
}

type storageStub struct{ data []byte }

func (s *storageStub) Put(_ context.Context, object ports.StoredObject) error {
	data, err := io.ReadAll(object.Reader)
	s.data = data
	return err
}
func (s *storageStub) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(string(s.data))), nil
}
func (s *storageStub) Delete(context.Context, string) error { return nil }

type speechStub struct{ err error }

func (s speechStub) Transcribe(_ context.Context, reader io.Reader, _ string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	data, err := io.ReadAll(reader)
	return string(data), err
}

type retryRepository struct{ *repositoryStub }

func (r *retryRepository) RetryProcessing(context.Context, int64, uuid.UUID, uuid.UUID, time.Time) (uuid.UUID, string, string, string, error) {
	return uuid.New(), "", "examinations/test/source.ogg", "voice.ogg", nil
}

type fileNameSpeechStub struct{ received chan string }

func (s fileNameSpeechStub) Transcribe(_ context.Context, reader io.Reader, fileName string) (string, error) {
	s.received <- fileName
	_, err := io.ReadAll(reader)
	return "тестовая транскрипция", err
}

type llmStub struct{}

func (llmStub) Summarize(context.Context, string) (string, error)      { return "summary", nil }
func (llmStub) Answer(context.Context, string, string) (string, error) { return "answer", nil }

func TestRetryRestoresOriginalFileName(t *testing.T) {
	t.Parallel()

	baseRepo := &repositoryStub{completed: make(chan struct{}, 1), failed: make(chan struct{}, 1)}
	repo := &retryRepository{repositoryStub: baseRepo}
	speech := fileNameSpeechStub{received: make(chan string, 1)}
	service := NewService(context.Background(), repo, speech, llmStub{}, &storageStub{data: []byte("audio")}, slog.Default(), 1, 1)
	defer service.Close()

	err := service.Retry(context.Background(), ports.RetryExaminationCommand{
		DoctorID:      42,
		ExaminationID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("повторить обработку: %v", err)
	}

	select {
	case fileName := <-speech.received:
		if fileName != "voice.ogg" {
			t.Fatalf("имя файла=%q, ожидалось %q", fileName, "voice.ogg")
		}
	case <-time.After(time.Second):
		t.Fatal("Speech-клиент не был вызван")
	}
}

func TestLoadProcessesInBackground(t *testing.T) {
	repo := &repositoryStub{completed: make(chan struct{}, 1), failed: make(chan struct{}, 1)}
	service := NewService(context.Background(), repo, speechStub{}, llmStub{}, &storageStub{}, slog.Default(), 1, 1)
	defer service.Close()
	id, err := service.Load(context.Background(), ports.LoadExaminationCommand{DoctorID: 42, FileName: "test.txt", File: io.NopCloser(strings.NewReader("patient transcript"))})
	if err != nil || id == uuid.Nil {
		t.Fatalf("Load(): идентификатор=%s, ошибка=%v", id, err)
	}
	select {
	case <-repo.completed:
	case <-time.After(time.Second):
		t.Fatal("обработка не завершилась")
	}
}

func TestLoadPersistsExternalClientFailure(t *testing.T) {
	repo := &repositoryStub{completed: make(chan struct{}, 1), failed: make(chan struct{}, 1)}
	service := NewService(context.Background(), repo, speechStub{err: errors.New("сервис распознавания недоступен")}, llmStub{}, &storageStub{}, slog.Default(), 1, 1)
	defer service.Close()
	_, err := service.Load(context.Background(), ports.LoadExaminationCommand{DoctorID: 42, File: io.NopCloser(strings.NewReader("audio"))})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-repo.failed:
	case <-time.After(time.Second):
		t.Fatal("ошибка обработки не была сохранена")
	}
}

func TestWorkerPoolRejectsTaskWhenBoundedQueueIsFull(t *testing.T) {
	baseRepo := &repositoryStub{completed: make(chan struct{}, 2), failed: make(chan struct{}, 2)}
	repo := &blockingRepository{
		repositoryStub: baseRepo,
		started:        make(chan struct{}),
		release:        make(chan struct{}),
	}
	service := NewService(context.Background(), repo, speechStub{}, llmStub{}, &storageStub{}, slog.Default(), 1, 1)
	defer service.Close()
	transcript := "тестовая транскрипция"
	newTask := func() processingTask {
		return processingTask{examinationID: uuid.New(), jobID: uuid.New(), transcript: &transcript}
	}

	if err := service.enqueue(newTask()); err != nil {
		t.Fatalf("первое задание не принято: %v", err)
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("обработчик не начал работу")
	}
	if err := service.enqueue(newTask()); err != nil {
		t.Fatalf("задание не добавлено в свободную очередь: %v", err)
	}
	if err := service.enqueue(newTask()); !errors.Is(err, ports.ErrProcessingQueueFull) {
		t.Fatalf("ожидалась ошибка заполненной очереди, получено: %v", err)
	}
	close(repo.release)
}

func TestServiceRestoresPendingTaskAfterStartup(t *testing.T) {
	examinationID, jobID := uuid.New(), uuid.New()
	repo := &repositoryStub{
		completed: make(chan struct{}, 1),
		failed:    make(chan struct{}, 1),
		pending: []ports.PendingProcessingTask{{
			JobID:         jobID,
			ExaminationID: examinationID,
			Attempt:       1,
			Transcript:    "сохранённая транскрипция",
		}},
	}
	service := NewService(context.Background(), repo, speechStub{}, llmStub{}, &storageStub{}, slog.Default(), 1, 1)
	defer service.Close()

	select {
	case <-repo.completed:
	case <-time.After(time.Second):
		t.Fatal("восстановленное задание не завершено")
	}
}

func TestWorkerPoolLimitsMaximumParallelism(t *testing.T) {
	const (
		workers   = 3
		taskCount = 12
	)
	baseRepo := &repositoryStub{
		completed: make(chan struct{}, taskCount),
		failed:    make(chan struct{}, taskCount),
	}
	repo := &concurrencyRepository{
		repositoryStub: baseRepo,
		started:        make(chan struct{}, taskCount),
		release:        make(chan struct{}),
	}
	service := NewService(context.Background(), repo, speechStub{}, llmStub{}, &storageStub{}, slog.Default(), workers, taskCount)
	var releaseOnce sync.Once
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(repo.release) })
		if err := service.Close(); err != nil {
			t.Errorf("ошибка Close(): %v", err)
		}
	})

	transcript := "тестовая транскрипция"
	for range taskCount {
		err := service.enqueue(processingTask{
			examinationID: uuid.New(),
			jobID:         uuid.New(),
			transcript:    &transcript,
		})
		if err != nil {
			t.Fatalf("задание не принято: %v", err)
		}
	}

	for range workers {
		select {
		case <-repo.started:
		case <-time.After(time.Second):
			t.Fatal("не все обработчики начали работу")
		}
	}
	if got := repo.maximum.Load(); got != workers {
		t.Fatalf("максимальный параллелизм = %d, ожидалось %d", got, workers)
	}

	releaseOnce.Do(func() { close(repo.release) })
	for range taskCount {
		select {
		case <-repo.completed:
		case <-time.After(time.Second):
			t.Fatal("не все задания завершились")
		}
	}
	if got := repo.maximum.Load(); got > workers {
		t.Fatalf("предел параллелизма нарушен: %d > %d", got, workers)
	}
}
