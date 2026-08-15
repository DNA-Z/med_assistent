package ports

import (
	"context"

	dto2 "github.com/DNA-Z/med_assistent/internal/application/use_cases/query/_dto"
	"github.com/google/uuid"
)

type ExaminationRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*dto2.ExaminationListItem, error)
	ListByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]*dto2.ExaminationDetails, error)
	FindByDoctorID(ctx context.Context, doctorID uuid.UUID, keyword string) ([]*dto2.ExaminationListItem, error)
}
