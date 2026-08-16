package ports

import (
	"context"
	"io"
)

type SpeechClient interface {
	Transcribe(ctx context.Context, audio io.Reader) (string, error)
}
