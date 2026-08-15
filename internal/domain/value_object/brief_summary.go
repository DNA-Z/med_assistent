package value_object

import (
	"errors"
	"strings"
)

var ErrEmptyBriefSummary = errors.New("brief summary cannot be empty")

// BriefSummary - краткая выжимка из транскрипции, содержащая ключевые моменты в нарушениях речи пациента
type BriefSummary struct {
	value string
}

func NewBriefSummary(value string) (BriefSummary, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return BriefSummary{}, ErrEmptyBriefSummary
	}

	return BriefSummary{
		value: value,
	}, nil
}

func (s BriefSummary) Value() string {
	return s.value
}

func (s BriefSummary) String() string {
	return s.value
}

func (s BriefSummary) IsEmpty() bool {
	return s.value == ""
}
