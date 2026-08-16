// internal/application/auth/service.go
package auth

import (
	"context"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type Service struct {
	doctors ports.DoctorWriteRepository
}

func NewService(
	doctors ports.DoctorWriteRepository,
) *Service {
	return &Service{
		doctors: doctors,
	}
}

func (s *Service) Start(
	ctx context.Context,
	doctorID int64,
) error {
	exists, err := s.doctors.Exists(ctx, doctorID)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return s.doctors.Create(ctx, doctorID)
}
