package query

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (s *Service) Get(ctx context.Context, q ports.GetExaminationQuery) (*ports.ExaminationDTO, error) {
	return s.readRepo.Get(ctx, q.DoctorID, q.ExaminationID)
}
