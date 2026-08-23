// Package openaiadapter реализует LLM-порт через OpenAI Responses API.
package openaiadapter

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

const defaultModel = "gpt-5.6"

// Client обращается к OpenAI Responses API через официальный Go SDK.
type Client struct {
	sdk   openai.Client
	model shared.ResponsesModel
}

// NewClient создаёт OpenAI-адаптер. Пустая модель заменяется на gpt-5.6,
// а baseURL можно передать для тестового сервера или совместимого API.
func NewClient(httpClient *http.Client, apiKey, model, baseURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if strings.TrimSpace(model) == "" {
		model = defaultModel
	}
	options := []option.RequestOption{
		option.WithHTTPClient(httpClient),
		option.WithAPIKey(apiKey),
	}
	if strings.TrimSpace(baseURL) != "" {
		options = append(options, option.WithBaseURL(baseURL))
	}
	return &Client{sdk: openai.NewClient(options...), model: shared.ResponsesModel(model)}
}

func (c *Client) ask(ctx context.Context, prompt string) (string, error) {
	response, err := c.sdk.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		// Медицинские материалы не должны сохраняться на стороне API без
		// отдельного осознанного решения владельца системы.
		Store: openai.Bool(false),
	})
	if err != nil {
		return "", fmt.Errorf("вызвать OpenAI Responses API: %w", err)
	}
	text := strings.TrimSpace(response.OutputText())
	if text == "" {
		return "", fmt.Errorf("OpenAI вернул пустой ответ")
	}
	return text, nil
}

// Summarize создаёт краткую медицинскую выжимку транскрипции.
func (c *Client) Summarize(ctx context.Context, transcript string) (string, error) {
	return c.ask(ctx, "Сделай краткую медицинскую выжимку следующего опроса, опираясь на нарушения в речи, интонацию, настроение пациента, - всего того, что может указывать на тот или иной диагноз. Не добавляй факты, которых нет в исходном тексте:\n"+transcript)
}

// Answer отвечает на вопрос врача только по переданным материалам обследований.
func (c *Client) Answer(ctx context.Context, contextText, question string) (string, error) {
	return c.ask(ctx, "Ответь на вопрос врача только по материалам опросов. Если данных недостаточно, прямо сообщи об этом.\nМатериалы опросов:\n"+contextText+"\nВопрос врача: "+question)
}

var _ ports.LLMClient = (*Client)(nil)
