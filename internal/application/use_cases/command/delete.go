package command

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/google/uuid"
)

func (s *Service) Delete(ctx context.Context, cmd ports.DeleteExaminationCommand) error {
	if cmd.DoctorID == 0 || cmd.ExaminationID == uuid.Nil {
		return ports.ErrInvalidCommand
	}
	if err := s.writeRepo.DeleteExamination(ctx, cmd.DoctorID, cmd.ExaminationID); err != nil {
		return err
	}
	s.logger.Info("обследование удалено", "doctor_id", cmd.DoctorID, "examination_id", cmd.ExaminationID)
	return nil
}
