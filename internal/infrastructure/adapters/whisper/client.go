package whisper

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

const defaultBaseURL = "https://api.whisper-api.com"

const defaultPollInterval = 2 * time.Second

const demoDelay = 5 * time.Second

//go:embed transcription.txt
var demoTranscript string

var _ ports.SpeechClient = (*Client)(nil)

type Client struct {
	httpClient   *http.Client
	apiKey       string
	baseURL      string
	pollInterval time.Duration
	transcript   string
	delay        time.Duration
}

func NewClient(
	httpClient *http.Client,
	apiKey string,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient:   httpClient,
		apiKey:       apiKey,
		baseURL:      defaultBaseURL,
		pollInterval: defaultPollInterval,
		transcript:   strings.TrimSpace(demoTranscript),
		delay:        demoDelay,
	}
}

func (c *Client) Transcribe(
	ctx context.Context,
	audio io.Reader,
	fileName string,
) (string, error) {
	if c.transcript != "" {
		timer := time.NewTimer(c.delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("имитация распознавания речи прервана: %w", ctx.Err())
		case <-timer.C:
			return c.transcript, nil
		}
	}

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

	if result, done, err := transcriptionResult(response); done {
		return result, err
	}
	if strings.TrimSpace(response.TaskID) == "" {
		return "", fmt.Errorf("Whisper API не вернул идентификатор задания")
	}

	return c.waitForResult(ctx, response.TaskID)
}

// waitForResult ожидает завершения асинхронного задания распознавания.
func (c *Client) waitForResult(ctx context.Context, taskID string) (string, error) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("ожидать транскрипцию Whisper API: %w", ctx.Err())
		case <-ticker.C:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/status/"+taskID, nil)
		if err != nil {
			return "", fmt.Errorf("создать запрос статуса Whisper API: %w", err)
		}
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("получить статус Whisper API: %w", err)
		}

		var status transcriptionResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&status)
		resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return "", fmt.Errorf("Whisper API вернул статус %d при проверке задания", resp.StatusCode)
		}
		if decodeErr != nil {
			return "", fmt.Errorf("декодировать статус Whisper API: %w", decodeErr)
		}

		if result, done, err := transcriptionResult(status); done {
			return result, err
		}
	}
}

// transcriptionResult интерпретирует текущее состояние задания Whisper API.
func transcriptionResult(response transcriptionResponse) (string, bool, error) {
	switch strings.ToLower(strings.TrimSpace(response.Status)) {
	case "queued", "pending", "processing", "running":
		return "", false, nil
	case "failed", "error", "cancelled", "canceled":
		if strings.TrimSpace(response.Error) != "" {
			return "", true, fmt.Errorf("Whisper API завершил задание с ошибкой: %s", response.Error)
		}
		return "", true, fmt.Errorf("Whisper API завершил задание со статусом %q", response.Status)
	case "completed", "complete", "succeeded", "success":
		if strings.TrimSpace(response.Result) == "" {
			return "", true, fmt.Errorf("Whisper API завершил задание с пустой транскрипцией")
		}
		return response.Result, true, nil
	default:
		return "", true, fmt.Errorf("Whisper API вернул неизвестный статус %q", response.Status)
	}
}

type transcriptionResponse struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`
	Result   string `json:"result"`
	Language string `json:"language"`
	Format   string `json:"format"`
	Error    string `json:"error"`
}
