package value_object

import (
	"errors"
)

var ErrInvalidDiagnosis = errors.New("некорректный диагноз")

// Diagnosis — тип для хранения кода диагноза по МКБ-11.
type Diagnosis struct {
	code        ICDCode
	description string
}

func NewDiagnosis(
	code ICDCode,
	description string,
) (Diagnosis, error) {
	if code.IsEmpty() {
		return Diagnosis{}, ErrInvalidDiagnosis
	}

	if description == "" {
		return Diagnosis{}, ErrInvalidDiagnosis
	}

	return Diagnosis{
		code:        code,
		description: description,
	}, nil
}

func (d Diagnosis) Code() ICDCode {
	return d.code
}

func (d Diagnosis) Description() string {
	return d.description
}

func (d Diagnosis) String() string {
	return d.code.String() + " - " + d.description
}
