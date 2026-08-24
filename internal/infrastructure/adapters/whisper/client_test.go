package whisper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTranscribeWaitsForCompletedTask(t *testing.T) {
	t.Parallel()

	var statusRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/transcribe":
			_, _ = io.WriteString(w, `{"task_id":"task-1","status":"queued","result":null}`)
		case r.Method == http.MethodGet && r.URL.Path == "/status/task-1":
			statusRequests.Add(1)
			_, _ = io.WriteString(w, `{"task_id":"task-1","status":"completed","result":"текст пациента"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.Client(), "test-key")
	client.transcript = ""
	client.baseURL = server.URL
	client.pollInterval = time.Millisecond

	result, err := client.Transcribe(context.Background(), strings.NewReader("audio"), "audio.ogg")
	if err != nil {
		t.Fatalf("распознать аудио: %v", err)
	}
	if result != "текст пациента" {
		t.Fatalf("получена транскрипция %q", result)
	}
	if statusRequests.Load() != 1 {
		t.Fatalf("количество запросов статуса: %d", statusRequests.Load())
	}
}

func TestTranscribeReturnsRemoteTaskError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_, _ = io.WriteString(w, `{"task_id":"task-2","status":"queued","result":null}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"task_id":"task-2","status":"failed","error":"invalid audio"}`)
	}))
	defer server.Close()

	client := NewClient(server.Client(), "test-key")
	client.transcript = ""
	client.baseURL = server.URL
	client.pollInterval = time.Millisecond

	_, err := client.Transcribe(context.Background(), strings.NewReader("audio"), "audio.ogg")
	if err == nil || !strings.Contains(err.Error(), "invalid audio") {
		t.Fatalf("ожидалась ошибка Whisper API, получено: %v", err)
	}
}

func TestTranscribeReturnsEmbeddedTranscriptAfterDelay(t *testing.T) {
	t.Parallel()

	client := NewClient(nil, "")
	client.delay = time.Millisecond

	result, err := client.Transcribe(context.Background(), strings.NewReader("любое аудио"), "voice.ogg")
	if err != nil {
		t.Fatalf("получить демонстрационную транскрипцию: %v", err)
	}
	if result != strings.TrimSpace(demoTranscript) {
		t.Fatal("клиент вернул текст, отличный от встроенной транскрипции")
	}
}

func TestTranscribeDemoRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	client := NewClient(nil, "")
	client.delay = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Transcribe(ctx, strings.NewReader("audio"), "voice.ogg"); err == nil {
		t.Fatal("ожидалась ошибка отмены контекста")
	}
}
