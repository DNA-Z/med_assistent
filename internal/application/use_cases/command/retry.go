package command

import (
	"context"
	"strings"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/google/uuid"
)

func (s *Service) Retry(ctx context.Context, cmd ports.RetryExaminationCommand) error {
	if cmd.DoctorID == 0 || cmd.ExaminationID == uuid.Nil {
		return ports.ErrInvalidCommand
	}
	jobID, transcript, err := s.writeRepo.RetryProcessing(ctx, cmd.DoctorID, cmd.ExaminationID, uuid.New(), time.Now().UTC())
	if err != nil {
		return err
	}
	if strings.TrimSpace(transcript) == "" {
		return ports.ErrFileRequired
	}
	s.logger.Info("повторная обработка поставлена в очередь", "doctor_id", cmd.DoctorID, "examination_id", cmd.ExaminationID, "job_id", jobID)
	s.processing.Go(func() error {
		s.process(cmd.ExaminationID, jobID, nil, "", &transcript)
		return nil
	})
	return nil
}
