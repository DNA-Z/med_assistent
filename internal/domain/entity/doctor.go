package entity

import (
	"errors"
	"strings"

	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
)

var (
	ErrInvalidDoctorID       = errors.New("invalid doctor id")
	ErrEmptyDoctorName       = errors.New("doctor name cannot be empty")
	ErrEmptyDoctorLastName   = errors.New("doctor last name cannot be empty")
	ErrEmptyDoctorPosition   = errors.New("doctor position cannot be empty")
	ErrInvalidWorkExperience = errors.New("invalid work experience")
)

// Doctor — врач-психиатр, работающий с системой.
type Doctor struct {
	id             uuid.UUID
	name           string
	lastName       string
	dateOfBirth    value_object.DateOfBirth
	position       string
	speciality     value_object.Speciality
	category       value_object.Category
	workExperience int
}

func NewDoctor(
	id uuid.UUID,
	name string,
	lastName string,
	dateOfBirth value_object.DateOfBirth,
	position string,
	speciality value_object.Speciality,
	category value_object.Category,
	workExperience int,
) (*Doctor, error) {

	if id == uuid.Nil {
		return nil, ErrInvalidDoctorID
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyDoctorName
	}

	lastName = strings.TrimSpace(lastName)
	if lastName == "" {
		return nil, ErrEmptyDoctorLastName
	}

	if dateOfBirth.IsZero() {
		return nil, value_object.ErrEmptyDateOfBirth
	}

	position = strings.TrimSpace(position)
	if position == "" {
		return nil, ErrEmptyDoctorPosition
	}

	if !speciality.IsValid() {
		return nil, value_object.ErrInvalidSpeciality
	}

	if !category.IsValid() {
		return nil, value_object.ErrInvalidCategory
	}

	if workExperience < 0 {
		return nil, ErrInvalidWorkExperience
	}

	return &Doctor{
		id:             id,
		name:           name,
		lastName:       lastName,
		dateOfBirth:    dateOfBirth,
		position:       position,
		speciality:     speciality,
		category:       category,
		workExperience: workExperience,
	}, nil
}

func (d Doctor) ID() uuid.UUID {
	return d.id
}

func (d Doctor) Name() string {
	return d.name
}

func (d Doctor) LastName() string {
	return d.lastName
}

func (d Doctor) DateOfBirth() value_object.DateOfBirth {
	return d.dateOfBirth
}

func (d Doctor) Position() string {
	return d.position
}

func (d Doctor) Speciality() value_object.Speciality {
	return d.speciality
}

func (d Doctor) Category() value_object.Category {
	return d.category
}

func (d Doctor) WorkExperience() int {
	return d.workExperience
}
