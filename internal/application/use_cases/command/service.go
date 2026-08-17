package command

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/domain/aggregate"
	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
)

type Service struct {
	writeRepo ports.ExaminationWriteRepository
	speech    ports.SpeechClient
	llm       ports.LLMClient
	logger    *slog.Logger

	processingCtx    context.Context
	processingCancel context.CancelFunc

	sem chan struct{}
	wg  sync.WaitGroup
}

func NewService(
	parent context.Context,
	writeRepo ports.ExaminationWriteRepository,
	speech ports.SpeechClient,
	llm ports.LLMClient,
	logger *slog.Logger,
	maxParallel int,
) *Service {
	if maxParallel <= 0 {
		maxParallel = 1
	}

	ctx, cancel := context.WithCancel(parent)

	return &Service{
		writeRepo:        writeRepo,
		speech:           speech,
		llm:              llm,
		logger:           logger,
		processingCtx:    ctx,
		processingCancel: cancel,
		sem:              make(chan struct{}, maxParallel),
	}
}

func (s *Service) Close() {
	s.processingCancel()
	s.wg.Wait()
}

func (s *Service) Load(
	ctx context.Context,
	cmd ports.LoadExaminationCommand,
) (uuid.UUID, error) {
	if cmd.DoctorID == 0 {
		return uuid.Nil, ports.ErrDoctorNotFound
	}

	if cmd.PatientID == uuid.Nil {
		return uuid.Nil, ports.ErrPatientRequired
	}

	if cmd.File == nil && cmd.Transcript == nil {
		return uuid.Nil, ports.ErrFileRequired
	}

	now := time.Now().UTC()

	examinationID := uuid.New()
	jobID := uuid.New()

	examination, err := aggregate.NewExamination(
		examinationID,
		now,
		uuid.NewSHA1(uuid.Nil, []byte(stringID(cmd.DoctorID))),
		cmd.PatientID,
	)
	if err != nil {
		return uuid.Nil, err
	}

	job, err := aggregate.NewProcessingJob(
		jobID,
		examinationID,
	)
	if err != nil {
		return uuid.Nil, err
	}

	err = s.writeRepo.CreateExamination(
		ctx,
		ports.ExaminationWriteModel{
			ID:              examination.ID(),
			DoctorID:        cmd.DoctorID,
			PatientID:       cmd.PatientID,
			ExaminationDate: examination.ExaminationDate(),
			Status:          examination.Status().String(),
			CreatedAt:       examination.CreatedAt(),
			UpdatedAt:       examination.UpdatedAt(),
		},
		ports.ProcessingJobWriteModel{
			ID:            job.ID(),
			ExaminationID: job.ExaminationID(),
			Status:        job.Status().String(),
			Attempt:       job.Attempt(),
			CreatedAt:     job.CreatedAt(),
			UpdatedAt:     job.UpdatedAt(),
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	file := cmd.File
	transcript := cmd.Transcript

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		if file != nil {
			defer file.Close()
		}

		s.process(
			examinationID,
			jobID,
			file,
			cmd.FileName,
			transcript,
		)
	}()

	return examinationID, nil
}

func (s *Service) process(
	examinationID uuid.UUID,
	jobID uuid.UUID,
	file io.Reader,
	fileName string,
	transcript *string,
) {
	select {
	case s.sem <- struct{}{}:
	case <-s.processingCtx.Done():
		return
	}
	defer func() {
		<-s.sem
	}()

	ctx, cancel := context.WithTimeout(
		s.processingCtx,
		15*time.Minute,
	)
	defer cancel()

	now := time.Now().UTC()

	if err := s.writeRepo.StartProcessing(
		ctx,
		examinationID,
		jobID,
		now,
		1,
	); err != nil {
		s.logger.Error(
			"failed to start processing",
			"examination_id", examinationID,
			"error", err,
		)
		return
	}

	var text string

	if transcript != nil {
		text = *transcript
	} else {
		var err error

		text, err = s.speech.Transcribe(
			ctx,
			file,
			fileName,
		)
		if err != nil {
			s.fail(examinationID, jobID, err)
			return
		}
	}

	if text == "" {
		s.fail(
			examinationID,
			jobID,
			errors.New("empty transcript"),
		)
		return
	}

	if err := s.writeRepo.SaveTranscript(
		ctx,
		examinationID,
		jobID,
		text,
		time.Now().UTC(),
	); err != nil {
		s.fail(examinationID, jobID, err)
		return
	}

	summary, err := s.llm.Summarize(ctx, text)
	if err != nil {
		s.fail(examinationID, jobID, err)
		return
	}

	if err := s.writeRepo.SaveSummary(
		ctx,
		examinationID,
		summary,
		time.Now().UTC(),
	); err != nil {
		s.fail(examinationID, jobID, err)
		return
	}

	if err := s.writeRepo.CompleteProcessing(
		ctx,
		examinationID,
		jobID,
		time.Now().UTC(),
	); err != nil {
		s.logger.Error(
			"failed to complete processing",
			"examination_id", examinationID,
			"error", err,
		)
	}
}

func (s *Service) fail(
	examinationID uuid.UUID,
	jobID uuid.UUID,
	err error,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if saveErr := s.writeRepo.FailProcessing(
		ctx,
		examinationID,
		jobID,
		err.Error(),
		time.Now().UTC(),
	); saveErr != nil {
		s.logger.Error(
			"failed to save processing error",
			"examination_id", examinationID,
			"error", saveErr,
		)
	}
}

func (s *Service) Retry(
	ctx context.Context,
	cmd ports.RetryExaminationCommand,
) error {
	if cmd.DoctorID == 0 || cmd.ExaminationID == uuid.Nil {
		return ports.ErrInvalidCommand
	}

	return s.writeRepo.RetryProcessing(
		ctx,
		cmd.ExaminationID,
		uuid.New(),
		time.Now().UTC(),
	)
}

func (s *Service) Delete(
	ctx context.Context,
	cmd ports.DeleteExaminationCommand,
) error {
	if cmd.DoctorID == 0 || cmd.ExaminationID == uuid.Nil {
		return ports.ErrInvalidCommand
	}

	return s.writeRepo.DeleteExamination(
		ctx,
		cmd.DoctorID,
		cmd.ExaminationID,
	)
}

func stringID(id int64) string {
	return time.Unix(id, 0).UTC().Format(time.RFC3339Nano)
}

var _ ports.ExaminationCommandHandler = (*Service)(nil)
var _ = value_object.ProcessingCreated
