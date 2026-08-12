package entity

import (
	"github.com/DNA-Z/med_assistent/internal/domain/shared"
	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
)

// Doctor - тип врач-психиатр
type Doctor struct {
	Name           string
	LastName       string
	YearOfBirth    shared.DateOfBirth
	Position       string
	Speciality     value_object.Speciality
	Category       value_object.Category
	WorkExperience int32
}
