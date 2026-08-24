package gigachat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

const defaultBaseURL = "https://api.giga.chat/v1"

var _ ports.LLMClient = (*Client)(nil)

type Client struct {
	httpClient *http.Client
	token      string
	model      string
	baseURL    string
}

func NewClient(
	httpClient *http.Client,
	token string,
	model string,
	baseURL string,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if model == "" {
		model = "GigaChat-2"
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		httpClient: httpClient,
		token:      token,
		model:      model,
		baseURL:    baseURL,
	}
}

func (c *Client) Ask(
	ctx context.Context,
	prompt string,
) (string, error) {
	requestBody := chatRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream:            false,
		RepetitionPenalty: 1,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("сериализовать запрос к GigaChat: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("создать запрос к GigaChat: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("выполнить запрос к GigaChat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf(
			"GigaChat API вернул статус %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var response chatResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("декодировать ответ GigaChat: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("GigaChat не вернул вариантов ответа")
	}

	content := response.Choices[0].Message.Content

	if content == "" {
		return "", fmt.Errorf("GigaChat вернул пустой ответ")
	}

	return content, nil
}

func (c *Client) Summarize(ctx context.Context, transcript string) (string, error) {
	return c.Ask(ctx, "Сделай краткую медицинскую выжимку следующего опроса:\n"+transcript)
}

func (c *Client) Answer(ctx context.Context, contextText, question string) (string, error) {
	return c.Ask(ctx, "Материалы опросов:\n"+contextText+"\nВопрос врача: "+question)
}

type chatRequest struct {
	Model             string    `json:"model"`
	Messages          []message `json:"messages"`
	Stream            bool      `json:"stream"`
	RepetitionPenalty float64   `json:"repetition_penalty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}
