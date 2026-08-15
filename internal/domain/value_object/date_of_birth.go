package value_object

import (
	"errors"
	"time"
)

var (
	ErrEmptyDateOfBirth  = errors.New("date of birth is empty")
	ErrFutureDateOfBirth = errors.New("date of birth cannot be in the future")
)

// DateOfBirth представляет дату рождения
type DateOfBirth struct {
	value time.Time
}

func NewDateOfBirth(value time.Time) (DateOfBirth, error) {
	if value.IsZero() {
		return DateOfBirth{}, ErrEmptyDateOfBirth
	}

	today := time.Now().UTC()

	birthDate := time.Date(
		value.Year(),
		value.Month(),
		value.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	todayDate := time.Date(
		today.Year(),
		today.Month(),
		today.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	if birthDate.After(todayDate) {
		return DateOfBirth{}, ErrFutureDateOfBirth
	}

	return DateOfBirth{
		value: birthDate,
	}, nil
}

func NewDateOfBirthFromString(value string) (DateOfBirth, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return DateOfBirth{}, err
	}

	return NewDateOfBirth(date)
}

func (d DateOfBirth) Value() time.Time {
	return d.value
}

func (d DateOfBirth) String() string {
	return d.value.Format("2006-01-02")
}

func (d DateOfBirth) Age(now time.Time) int {
	age := now.Year() - d.value.Year()

	birthday := time.Date(
		now.Year(),
		d.value.Month(),
		d.value.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	if now.Before(birthday) {
		age--
	}

	return age
}

func (d DateOfBirth) IsZero() bool {
	return d.value.IsZero()
}
