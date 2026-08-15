package dto

import (
	"time"

	"github.com/google/uuid"
)

type ExaminationDetails struct {
	ID            uuid.UUID
	ExaminationAt time.Time

	PatientID       uuid.UUID
	PatientName     string
	PatientLastName string

	Transcript string
	Summary    string
	Diagnosis  string

	Status       string
	ErrorMessage string
}
