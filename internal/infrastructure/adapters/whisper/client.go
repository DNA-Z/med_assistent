package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

const defaultBaseURL = "https://api.whisper-api.com"

var _ ports.SpeechClient = (*Client)(nil)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewClient(
	httpClient *http.Client,
	apiKey string,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient: httpClient,
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
	}
}

func (c *Client) Transcribe(
	ctx context.Context,
	audio io.Reader,
	fileName string,
) (string, error) {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	if fileName == "" {
		fileName = "audio.mp3"
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("создать multipart-часть файла: %w", err)
	}

	if _, err := io.Copy(part, audio); err != nil {
		return "", fmt.Errorf("скопировать аудио: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("закрыть multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/transcribe",
		&body,
	)
	if err != nil {
		return "", fmt.Errorf("создать запрос к Whisper API: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("выполнить запрос к Whisper API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf(
			"Whisper API вернул статус %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var response transcriptionResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("декодировать ответ Whisper API: %w", err)
	}

	if response.Result == "" {
		return "", fmt.Errorf("Whisper API вернул пустую транскрипцию")
	}

	return response.Result, nil
}

type transcriptionResponse struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`
	Result   string `json:"result"`
	Language string `json:"language"`
	Format   string `json:"format"`
}
