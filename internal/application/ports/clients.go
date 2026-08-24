package ports

import (
	"context"
	"io"
)

// StoredObject описывает сохраняемый объект без привязки к конкретному S3-провайдеру.
type StoredObject struct {
	Key         string
	Reader      io.Reader
	Size        int64
	ContentType string
}

// ObjectStorage задаёт порт постоянного хранения исходных файлов обследований.
type ObjectStorage interface {
	Put(ctx context.Context, object StoredObject) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

// SpeechClient абстрагирует поставщика распознавания речи.
type SpeechClient interface {
	Transcribe(
		ctx context.Context,
		file io.Reader,
		fileName string,
	) (string, error)
}

// LLMClient абстрагирует поставщика языковой модели.
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
