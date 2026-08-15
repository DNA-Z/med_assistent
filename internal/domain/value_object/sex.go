package value_object

import "errors"

type Sex string

// Константы пола
const (
	Male   Sex = "мужской"
	Female Sex = "женский"
)

var ErrInvalidSex = errors.New("invalid sex")

func NewSex(value string) (Sex, error) {
	sex := Sex(value)

	if !sex.IsValid() {
		return "", ErrInvalidSex
	}

	return sex, nil
}

func (s Sex) IsValid() bool {
	switch s {
	case Male, Female:
		return true
	default:
		return false
	}
}

func (s Sex) String() string {
	switch s {
	case Male:
		return "мужской"
	case Female:
		return "женский"
	default:
		return "неизвестный"
	}
}
