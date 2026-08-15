package auth

import (
	"context"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/domain/entity"
	"github.com/google/uuid"
)

type Service struct {
	doctors ports.DoctorRepository
}

func NewService(
	doctors ports.DoctorRepository,
) *Service {
	return &Service{
		doctors: doctors,
	}
}

type Identity struct {
	TelegramID int64
	Name       string
	LastName   string
}

func (s *Service) Authenticate(
	ctx context.Context,
	identity Identity,
) (*entity.Doctor, error) {

	doctor, err := s.doctors.GetByID(
		ctx,
		identity.TelegramID,
	)

	if err == nil {
		return doctor, nil
	}

	// Здесь в реальном приложении лучше использовать
	// отдельный registration flow.
	doctorID := uuid.New()

	doctor, err = entity.NewDoctor(
		doctorID,
		identity.Name,
		identity.LastName,
		// остальные необходимые данные
	)

	if err != nil {
		return nil, err
	}

	if err := s.doctors.Create(ctx, doctor); err != nil {
		return nil, err
	}

	return doctor, nil
}
