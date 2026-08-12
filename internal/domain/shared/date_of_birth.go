package shared

import "time"

// DateOfBirth — значение-объект для представления даты рождения.
type DateOfBirth struct {
	value time.Time
}

// NewDateOfBirth создаёт DateOfBirth из time.Time.
func NewDateOfBirth(t time.Time) DateOfBirth {
	return DateOfBirth{value: t}
}

// FromString создаёт DateOfBirth из строки формата "2006-01-02".
func FromString(s string) (DateOfBirth, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return DateOfBirth{}, err
	}
	return DateOfBirth{value: t}, nil
}

func (d DateOfBirth) String() string {
	return d.value.Format("2006-01-02")
}
