package entity

import (
	"time"

	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
)

type Patient struct {
	ID          uuid.UUID
	Name        string
	LastName    string
	Sex         string
	Age         int
	YearOfBirth time.Time
	Diagnosis   value_object.Diagnosis
}
