package value_object

import (
	"errors"
	"strings"
)

var ErrInvalidICDCode = errors.New("некорректный код МКБ-11")

// ICDCode — код диагноза по МКБ-11.
type ICDCode string

func NewICDCode(value string) (ICDCode, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", ErrInvalidICDCode
	}

	return ICDCode(strings.ToUpper(value)), nil
}

func (c ICDCode) String() string {
	return string(c)
}

func (c ICDCode) IsEmpty() bool {
	return c == ""
}
