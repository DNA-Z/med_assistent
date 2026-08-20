package command

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

func (s *Service) process(examinationID, jobID uuid.UUID, file io.Reader, fileName string, transcript *string) {
	select {
	case s.sem <- struct{}{}:
	case <-s.processingCtx.Done():
		return
	}
	defer func() { <-s.sem }()
	ctx, cancel := context.WithTimeout(s.processingCtx, 15*time.Minute)
	defer cancel()
	if err := s.writeRepo.StartProcessing(ctx, examinationID, jobID, time.Now().UTC(), 1); err != nil {
		s.logger.Error("failed to start processing", "examination_id", examinationID, "error", err)
		return
	}
	var text string
	if transcript != nil {
		text = *transcript
	} else {
		var err error
		text, err = s.speech.Transcribe(ctx, file, fileName)
		if err != nil {
			s.fail(examinationID, jobID, err)
			return
		}
	}
	if text == "" {
		s.fail(examinationID, jobID, errors.New("empty transcript"))
		return
	}
	if err := s.writeRepo.SaveTranscript(ctx, examinationID, jobID, text, time.Now().UTC()); err != nil {
		s.fail(examinationID, jobID, err)
		return
	}
	summary, err := s.llm.Summarize(ctx, text)
	if err != nil {
		s.fail(examinationID, jobID, err)
		return
	}
	if err := s.writeRepo.SaveSummary(ctx, examinationID, summary, time.Now().UTC()); err != nil {
		s.fail(examinationID, jobID, err)
		return
	}
	if err := s.writeRepo.CompleteProcessing(ctx, examinationID, jobID, time.Now().UTC()); err != nil {
		s.logger.Error("failed to complete processing", "examination_id", examinationID, "error", err)
	}
}

func (s *Service) fail(examinationID, jobID uuid.UUID, processingErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.writeRepo.FailProcessing(ctx, examinationID, jobID, processingErr.Error(), time.Now().UTC()); err != nil {
		s.logger.Error("failed to save processing error", "examination_id", examinationID, "error", err)
	}
}
