package mock

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type SpeechClient struct{}

func (SpeechClient) Transcribe(ctx context.Context, file io.Reader, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("прочитать тестовую запись: %w", err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("тестовая запись пуста")
	}
	return text, nil
}

type LLMClient struct{}

func (LLMClient) Summarize(ctx context.Context, transcript string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	text := strings.TrimSpace(transcript)
	if len([]rune(text)) > 240 {
		text = string([]rune(text)[:240]) + "…"
	}
	return "Краткая выжимка: " + text, nil
}

func (LLMClient) Answer(ctx context.Context, contextText, question string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(contextText) == "" {
		return "", fmt.Errorf("отсутствует контекст обследований")
	}
	return "Тестовый ответ на вопрос «" + strings.TrimSpace(question) + "» по сохранённым материалам.", nil
}
