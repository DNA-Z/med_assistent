package dto

import (
	"time"

	"github.com/google/uuid"
)

type ExaminationListItem struct {
	ID            uuid.UUID
	ExaminationAt time.Time
	PatientName   string
	Summary       string
	Status        string
}
