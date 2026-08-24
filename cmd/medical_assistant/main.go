package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/auth"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/application/use_cases/command"
	queryapp "github.com/DNA-Z/med_assistent/internal/application/use_cases/query"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/gigachat"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/mock"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/objectstorage"
	openaiadapter "github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/openai"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/postgres"
	redisadapter "github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/redis"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/whisper"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/config"
	"github.com/DNA-Z/med_assistent/internal/presentation/adapters/telegram"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("приложение остановлено с ошибкой", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) (runErr error) {
	cfg := config.NewConfig()
	if err := cfg.ConfigInit(); err != nil {
		return fmt.Errorf("загрузить конфигурацию: %w", err)
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pg, err := postgres.New(ctx, cfg.DBConnectionString)
	if err != nil {
		return fmt.Errorf("подключиться к PostgreSQL: %w", err)
	}
	defer func() {
		pg.Close()
		logger.Info("соединение с PostgreSQL закрыто")
	}()

	rdb, err := redisadapter.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("подключиться к Redis: %w", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("закрыть соединение с Redis: %w", err))
			return
		}
		logger.Info("соединение с Redis закрыто")
	}()

	speech, llm, err := clients(cfg)
	if err != nil {
		return fmt.Errorf("создать клиенты внешних сервисов: %w", err)
	}
	storage, err := objectstorage.NewMinIO(ctx, objectstorage.Config{Endpoint: cfg.ObjectStorage.Endpoint, AccessKey: cfg.ObjectStorage.AccessKey, SecretKey: cfg.ObjectStorage.SecretKey, Bucket: cfg.ObjectStorage.Bucket, UseSSL: cfg.ObjectStorage.UseSSL})
	if err != nil {
		return fmt.Errorf("подключиться к объектному хранилищу: %w", err)
	}
	logger.Info("подключение к объектному хранилищу установлено", "bucket", cfg.ObjectStorage.Bucket)

	commands := command.NewService(ctx, postgres.NewExaminationWriteRepository(pg), speech, llm, storage, logger, cfg.Processing.Workers, cfg.Processing.QueueSize)
	defer func() {
		runErr = errors.Join(runErr, commands.Close())
		logger.Info("фоновые задачи обработки остановлены")
	}()
	queries := queryapp.NewService(redisadapter.NewExaminationReadRepository(rdb), llm, logger)
	authService := auth.NewService(postgres.NewDoctorRepository(pg), logger)

	outbox := redisadapter.NewOutboxWorker(pg.Pool(), rdb, logger, storage)
	bot, err := telegram.New(telegram.Config{Token: cfg.Telegram.Token, Timeout: cfg.Telegram.Timeout}, telegram.Dependencies{Auth: authService, Commands: commands, Queries: queries, Logger: logger})
	if err != nil {
		return fmt.Errorf("создать Telegram-бота: %w", err)
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		outbox.Run(groupCtx)
		return nil
	})
	group.Go(func() error {
		bot.Start()
		if groupCtx.Err() == nil {
			return errors.New("Telegram-бот неожиданно остановился")
		}
		return nil
	})
	group.Go(func() error {
		<-groupCtx.Done()
		logger.Info("получен сигнал завершения", "reason", groupCtx.Err())
		bot.Stop()
		return nil
	})

	logger.Info("приложение запущено", "workers", cfg.Processing.Workers, "queue_size", cfg.Processing.QueueSize)
	if err := waitForGroup(ctx, group, 10*time.Second); err != nil {
		stop()
		return fmt.Errorf("выполнить фоновые компоненты: %w", err)
	}
	logger.Info("приложение остановлено")
	return nil
}

func validateConfig(cfg *config.Config) error {
	if cfg.DBConnectionString == "" {
		return errors.New("не указана строка подключения к PostgreSQL")
	}
	if cfg.Redis.Address == "" {
		return errors.New("не указан адрес Redis")
	}
	if cfg.Telegram.Token == "" {
		return errors.New("не указана переменная TELEGRAM_TOKEN")
	}
	if cfg.ObjectStorage.Endpoint == "" || cfg.ObjectStorage.AccessKey == "" || cfg.ObjectStorage.SecretKey == "" || cfg.ObjectStorage.Bucket == "" {
		return errors.New("не заполнена конфигурация объектного хранилища")
	}
	if cfg.Processing.Workers <= 0 {
		return errors.New("количество обработчиков фоновых заданий должно быть положительным")
	}
	if cfg.Processing.QueueSize <= 0 {
		return errors.New("размер очереди обработки должен быть положительным")
	}
	return nil
}

func waitForGroup(ctx context.Context, group *errgroup.Group, shutdownTimeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- group.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		timer := time.NewTimer(shutdownTimeout)
		defer timer.Stop()
		select {
		case err := <-done:
			return err
		case <-timer.C:
			return errors.New("превышено время ожидания корректного завершения")
		}
	}
}

func clients(cfg *config.Config) (ports.SpeechClient, ports.LLMClient, error) {
	var speech ports.SpeechClient = mock.SpeechClient{}
	var llm ports.LLMClient = mock.LLMClient{}
	if cfg.Speech.Provider == "whisper" {
		speech = whisper.NewClient(&http.Client{Timeout: time.Duration(cfg.Speech.Timeout) * time.Second}, cfg.Speech.APIKey)
	} else if cfg.Speech.Provider != "mock" {
		return nil, nil, fmt.Errorf("неизвестный провайдер распознавания речи %q", cfg.Speech.Provider)
	}
	if cfg.LLM.Provider == "gigachat" {
		if cfg.LLM.GigaChat.Token == "" {
			return nil, nil, errors.New("для провайдера GigaChat не указана переменная GIGACHAT_TOKEN")
		}
		llm = gigachat.NewClient(&http.Client{Timeout: time.Duration(cfg.LLM.Timeout) * time.Second}, cfg.LLM.GigaChat.Token, cfg.LLM.GigaChat.Model, cfg.LLM.GigaChat.BaseURL)
	} else if cfg.LLM.Provider == "openai" {
		if cfg.LLM.OpenAI.APIKey == "" {
			return nil, nil, errors.New("для провайдера OpenAI не указана переменная OPENAI_API_KEY")
		}
		llm = openaiadapter.NewClient(&http.Client{Timeout: time.Duration(cfg.LLM.Timeout) * time.Second}, cfg.LLM.OpenAI.APIKey, cfg.LLM.OpenAI.Model, cfg.LLM.OpenAI.BaseURL)
	} else if cfg.LLM.Provider != "mock" {
		return nil, nil, fmt.Errorf("неизвестный провайдер языковой модели %q", cfg.LLM.Provider)
	}
	return speech, llm, nil
}
