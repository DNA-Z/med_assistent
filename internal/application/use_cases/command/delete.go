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
	return s.writeRepo.DeleteExamination(ctx, cmd.DoctorID, cmd.ExaminationID)
}
