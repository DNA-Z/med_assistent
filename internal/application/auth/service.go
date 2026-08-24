package auth

import (
	"context"
	"log/slog"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type Service struct {
	doctors ports.DoctorWriteRepository
	logger  *slog.Logger
}

func NewService(
	doctors ports.DoctorWriteRepository,
	logger *slog.Logger,
) *Service {
	return &Service{
		doctors: doctors,
		logger:  logger,
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
		s.logger.Debug("врач уже зарегистрирован", "doctor_id", doctorID)
		return nil
	}

	if err := s.doctors.Create(ctx, doctorID); err != nil {
		return err
	}
	s.logger.Info("врач зарегистрирован", "doctor_id", doctorID)
	return nil
}
