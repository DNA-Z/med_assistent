package value_object

import (
	"errors"
)

type Speciality string

const (
	Psychiatrist Speciality = "psychiatrist"
)

var ErrInvalidSpeciality = errors.New("некорректная специальность")

func NewSpeciality(value string) (Speciality, error) {
	speciality := Speciality(value)

	if !speciality.IsValid() {
		return "", ErrInvalidSpeciality
	}

	return speciality, nil
}

func (s Speciality) IsValid() bool {
	switch s {
	case Psychiatrist:
		return true
	default:
		return false
	}
}

func (s Speciality) String() string {
	switch s {
	case Psychiatrist:
		return "Врач-психиатр"
	default:
		return "Неизвестная специальность"
	}
}
