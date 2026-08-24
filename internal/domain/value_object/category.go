package value_object

import (
	"errors"
)

// Category — категория врача.
type Category int

const (
	FirstCategory Category = iota + 1
	SecondCategory
	HighestCategory
)

var ErrInvalidCategory = errors.New("некорректная категория врача")

func NewCategory(value int) (Category, error) {
	category := Category(value)

	if !category.IsValid() {
		return 0, ErrInvalidCategory
	}

	return category, nil
}

func (c Category) IsValid() bool {
	return c >= FirstCategory && c <= HighestCategory
}

func (c Category) String() string {
	switch c {
	case FirstCategory:
		return "первая"
	case SecondCategory:
		return "вторая"
	case HighestCategory:
		return "высшая"
	default:
		return "неизвестная"
	}
}
