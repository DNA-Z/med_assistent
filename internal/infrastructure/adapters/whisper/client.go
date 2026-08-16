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
) (string, error) {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return "", fmt.Errorf("create multipart file: %w", err)
	}

	if _, err := io.Copy(part, audio); err != nil {
		return "", fmt.Errorf("copy audio: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/transcribe",
		&body,
	)
	if err != nil {
		return "", fmt.Errorf("create whisper request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("whisper request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf(
			"whisper API returned status %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var response transcriptionResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode whisper response: %w", err)
	}

	if response.Result == "" {
		return "", fmt.Errorf("whisper returned empty transcription")
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
