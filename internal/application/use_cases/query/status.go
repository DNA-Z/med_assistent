package query

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (s *Service) Status(ctx context.Context, q ports.GetExaminationStatusQuery) (*ports.ExaminationStatusDTO, error) {
	return s.readRepo.Status(ctx, q.DoctorID, q.ExaminationID)
}
