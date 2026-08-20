package query

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"strings"
)

func (s *Service) Find(ctx context.Context, q ports.FindExaminationsQuery) ([]ports.ExaminationDTO, error) {
	if strings.TrimSpace(q.Keyword) == "" {
		return nil, ports.ErrEmptyKeyword
	}
	return s.readRepo.Find(ctx, q.DoctorID, q.Keyword)
}
