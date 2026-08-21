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
	s.logger.Info("обработка обследования начата", "examination_id", examinationID, "job_id", jobID, "status", "processing")
	var text string
	if transcript != nil {
		text = *transcript
	} else {
		var err error
		s.logger.Info("вызов Speech-клиента", "examination_id", examinationID, "file_name", fileName)
		text, err = s.speech.Transcribe(ctx, file, fileName)
		if err != nil {
			s.fail(examinationID, jobID, err)
			return
		}
		s.logger.Info("Speech-клиент завершил распознавание", "examination_id", examinationID)
	}
	if text == "" {
		s.fail(examinationID, jobID, errors.New("empty transcript"))
		return
	}
	if err := s.writeRepo.SaveTranscript(ctx, examinationID, jobID, text, time.Now().UTC()); err != nil {
		s.fail(examinationID, jobID, err)
		return
	}
	s.logger.Info("транскрипция сохранена", "examination_id", examinationID, "status", "transcribed")
	s.logger.Info("вызов LLM-клиента для создания выжимки", "examination_id", examinationID)
	summary, err := s.llm.Summarize(ctx, text)
	if err != nil {
		s.fail(examinationID, jobID, err)
		return
	}
	s.logger.Info("LLM-клиент создал выжимку", "examination_id", examinationID)
	if err := s.writeRepo.SaveSummary(ctx, examinationID, summary, time.Now().UTC()); err != nil {
		s.fail(examinationID, jobID, err)
		return
	}
	s.logger.Info("краткая выжимка сохранена", "examination_id", examinationID, "status", "summarized")
	if err := s.writeRepo.CompleteProcessing(ctx, examinationID, jobID, time.Now().UTC()); err != nil {
		s.logger.Error("failed to complete processing", "examination_id", examinationID, "error", err)
		return
	}
	s.logger.Info("обработка обследования завершена", "examination_id", examinationID, "job_id", jobID, "status", "completed")
}

func (s *Service) fail(examinationID, jobID uuid.UUID, processingErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.writeRepo.FailProcessing(ctx, examinationID, jobID, processingErr.Error(), time.Now().UTC()); err != nil {
		s.logger.Error("failed to save processing error", "examination_id", examinationID, "error", err)
		return
	}
	s.logger.Error("обработка обследования завершилась ошибкой", "examination_id", examinationID, "job_id", jobID, "status", "failed", "error", processingErr)
}
