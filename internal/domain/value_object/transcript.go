package value_object

import (
	"errors"
	"strings"
)

var ErrEmptyTranscript = errors.New("transcript cannot be empty")

type Transcript struct {
	value string
}

func NewTranscript(value string) (Transcript, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Transcript{}, ErrEmptyTranscript
	}

	return Transcript{
		value: value,
	}, nil
}

func (t Transcript) Value() string {
	return t.value
}

func (t Transcript) String() string {
	return t.value
}

func (t Transcript) IsEmpty() bool {
	return t.value == ""
}
