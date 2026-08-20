package command

import (
	"context"
	"log/slog"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"golang.org/x/sync/errgroup"
)

// Service is an application facade. Each command is implemented in its own file.
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

func NewService(parent context.Context, writeRepo ports.ExaminationWriteRepository, speech ports.SpeechClient, llm ports.LLMClient, logger *slog.Logger, maxParallel int) *Service {
	if maxParallel <= 0 {
		maxParallel = 1
	}
	ctx, cancel := context.WithCancel(parent)
	return &Service{writeRepo: writeRepo, speech: speech, llm: llm, logger: logger, processingCtx: ctx, processingCancel: cancel, sem: make(chan struct{}, maxParallel)}
}

func (s *Service) Close() error {
	s.processingCancel()
	return s.processing.Wait()
}

var _ ports.ExaminationCommandHandler = (*Service)(nil)
