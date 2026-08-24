package openaiadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientUsesResponsesAPI(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/responses" {
			t.Errorf("путь запроса = %q", request.URL.Path)
		}
		if authorization := request.Header.Get("Authorization"); authorization != "Bearer test-key" {
			t.Errorf("заголовок Authorization = %q", authorization)
		}
		var body struct {
			Model string `json:"model"`
			Input string `json:"input"`
			Store bool   `json:"store"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("не удалось декодировать запрос: %v", err)
		}
		if body.Model != "gpt-test" {
			t.Errorf("модель = %q", body.Model)
		}
		if !strings.Contains(body.Input, "жалобы пациента") {
			t.Errorf("в запросе отсутствует транскрипция: %q", body.Input)
		}
		if body.Store {
			t.Error("ответ не должен сохраняться на стороне API")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"resp_test","object":"response","model":"gpt-test","status":"completed","output":[{"id":"msg_test","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"  Краткая выжимка  ","annotations":[]}]}]}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), "test-key", "gpt-test", server.URL)
	result, err := client.Summarize(context.Background(), "жалобы пациента")
	if err != nil {
		t.Fatalf("ошибка Summarize(): %v", err)
	}
	if result != "Краткая выжимка" {
		t.Fatalf("результат = %q", result)
	}
}

func TestClientRejectsEmptyResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"resp_test","object":"response","model":"gpt-test","status":"completed","output":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), "test-key", "gpt-test", server.URL)
	if _, err := client.Answer(context.Background(), "контекст", "вопрос"); err == nil || !strings.Contains(err.Error(), "пустой ответ") {
		t.Fatalf("ожидалась ошибка пустого ответа, получено: %v", err)
	}
}
