package ports

import (
	"context"

	"github.com/DNA-Z/med_assistent/internal/domain/entity"
	"github.com/google/uuid"
)

type DoctorRepository interface {
	Create(ctx context.Context, doctor *entity.Doctor) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Doctor, error)
}
