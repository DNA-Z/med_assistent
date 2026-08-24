package query

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (s *Service) List(ctx context.Context, q ports.ListExaminationsQuery) ([]ports.ExaminationDTO, error) {
	return s.readRepo.List(ctx, q.DoctorID)
}
