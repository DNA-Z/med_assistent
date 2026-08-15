package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
)

var (
	ErrInvalidPatientID     = errors.New("invalid patient id")
	ErrEmptyPatientName     = errors.New("patient name cannot be empty")
	ErrEmptyPatientLastName = errors.New("patient last name cannot be empty")
)

type Patient struct {
	id          uuid.UUID
	name        string
	lastName    string
	sex         value_object.Sex
	dateOfBirth value_object.DateOfBirth
}

func NewPatient(
	id uuid.UUID,
	name string,
	lastName string,
	sex value_object.Sex,
	dateOfBirth value_object.DateOfBirth,
) (*Patient, error) {

	if id == uuid.Nil {
		return nil, ErrInvalidPatientID
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyPatientName
	}

	lastName = strings.TrimSpace(lastName)
	if lastName == "" {
		return nil, ErrEmptyPatientLastName
	}

	if !sex.IsValid() {
		return nil, value_object.ErrInvalidSex
	}

	if dateOfBirth.IsZero() {
		return nil, value_object.ErrEmptyDateOfBirth
	}

	return &Patient{
		id:          id,
		name:        name,
		lastName:    lastName,
		sex:         sex,
		dateOfBirth: dateOfBirth,
	}, nil
}

func (p Patient) ID() uuid.UUID {
	return p.id
}

func (p Patient) Name() string {
	return p.name
}

func (p Patient) LastName() string {
	return p.lastName
}

func (p Patient) Sex() value_object.Sex {
	return p.sex
}

func (p Patient) DateOfBirth() value_object.DateOfBirth {
	return p.dateOfBirth
}

func (p Patient) Age(now time.Time) int {
	return p.dateOfBirth.Age(now)
}
