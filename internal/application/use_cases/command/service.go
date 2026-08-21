package command

import (
	"context"
	"log/slog"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"golang.org/x/sync/errgroup"
)

// Service объединяет обработчики команд приложения и управляет фоновыми задачами.
// Реализация каждой команды вынесена в отдельный файл пакета command.
type Service struct {
	writeRepo ports.ExaminationWriteRepository
	speech    ports.SpeechClient
	llm       ports.LLMClient
	logger    *slog.Logger

	processingCtx    context.Context
	processingCancel context.CancelFunc
	sem              chan struct{}
	processing       errgroup.Group
}

// NewService создаёт фасад команд и ограничивает число одновременно
// обрабатываемых обследований значением maxParallel.
func NewService(parent context.Context, writeRepo ports.ExaminationWriteRepository, speech ports.SpeechClient, llm ports.LLMClient, logger *slog.Logger, maxParallel int) *Service {
	if maxParallel <= 0 {
		maxParallel = 1
	}
	ctx, cancel := context.WithCancel(parent)
	return &Service{writeRepo: writeRepo, speech: speech, llm: llm, logger: logger, processingCtx: ctx, processingCancel: cancel, sem: make(chan struct{}, maxParallel)}
}

// Close запрещает запуск новой фоновой работы, отменяет текущую и ожидает
// завершения всех зарегистрированных задач обработки.
func (s *Service) Close() error {
	s.processingCancel()
	return s.processing.Wait()
}

var _ ports.ExaminationCommandHandler = (*Service)(nil)
