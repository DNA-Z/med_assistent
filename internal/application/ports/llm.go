package ports

import "context"

type LLMClient interface {
	Ask(ctx context.Context, prompt string) (string, error)
}
