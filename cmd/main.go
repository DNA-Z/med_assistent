package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/DNA-Z/med_assistent/internal/application"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/gigachat"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/whisper"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/config"
	"github.com/DNA-Z/med_assistent/internal/presentation/adapters/telegram"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	httpClient := &http.Client{}

	speechClient := whisper.NewClient(
		httpClient,
		cfg.Whisper.APIKey,
	)

	llmClient := gigachat.NewClient(
		httpClient,
		cfg.GigaChat.Token,
		cfg.GigaChat.Model,
	)

	app := application.New(
		speechClient,
		llmClient,
		logger,
	)

	bot, err := telegram.New(
		cfg.Telegram.Token,
		app,
		logger,
	)
	if err != nil {
		log.Fatalf("create telegram bot: %v", err)
	}

	if err := bot.Start(); err != nil {
		log.Fatalf("telegram bot stopped: %v", err)
	}
}
