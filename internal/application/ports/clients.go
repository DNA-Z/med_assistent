package ports

import (
	"context"
	"io"
)

type SpeechClient interface {
	Transcribe(
		ctx context.Context,
		file io.Reader,
		fileName string,
	) (string, error)
}

type LLMClient interface {
	Summarize(
		ctx context.Context,
		transcript string,
	) (string, error)

	Answer(
		ctx context.Context,
		contextText string,
		question string,
	) (string, error)
}
