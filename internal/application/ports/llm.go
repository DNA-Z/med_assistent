package ports

import "context"

type LLMClient interface {
	Summarize(
		ctx context.Context,
		transcript string,
	) (string, error)

	Ask(
		ctx context.Context,
		contextText string,
		question string,
	) (string, error)
}
