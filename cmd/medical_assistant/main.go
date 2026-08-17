package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/auth"
	"github.com/DNA-Z/med_assistent/internal/application/use_cases/command"
	"github.com/DNA-Z/med_assistent/internal/application/use_cases/query"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/postgres"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/redis"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/config"
	"github.com/DNA-Z/med_assistent/internal/presentation/adapters/telegram"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	cfg := config.NewConfig()
	cfg.ConfigInit()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	pg, err := postgres.New(
		ctx,
		postgres.Config{
			DSN: cfg.DBConnectionString,
		},
	)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := pg.Close(); err != nil {
			logger.Error("failed to close postgres", "error", err)
		}
	}()

	redisClient, err := redis.New(
		ctx,
		redis.Config{
			Address:  cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
	)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("failed to close redis", "error", err)
		}
	}()

	doctorRepository := postgres.NewDoctorRepository(pg)
	patientRepository := postgres.NewPatientRepository(pg)
	examinationRepository := postgres.NewExaminationRepository(pg)
	processingJobRepository := postgres.NewProcessingJobRepository(pg)
	outboxRepository := postgres.NewOutboxRepository(pg)

	examinationReadRepository := redis.NewExaminationReadRepository(
		redisClient,
	)

	speechClient, err := speech.New(
		speech.Config{
			Provider: cfg.Speech.Provider,
			APIKey:   cfg.Speech.APIKey,
			BaseURL:  cfg.Speech.BaseURL,
			Timeout:  time.Duration(cfg.Speech.Timeout) * time.Second,
		},
	)
	if err != nil {
		logger.Error("failed to create speech client", "error", err)
		os.Exit(1)
	}

	authService := auth.NewService(
		doctorRepository,
		logger,
	)

	examinationCommandHandler := command.NewExaminationHandler(
		doctorRepository,
		patientRepository,
		examinationRepository,
		processingJobRepository,
		outboxRepository,
		speechClient,
		logger,
	)

	examinationQueryHandler := query.NewExaminationHandler(
		examinationReadRepository,
		logger,
	)

	telegramBot, err := telegram.New(
		telegram.Config{
			Token:   cfg.Telegram.Token,
			Timeout: cfg.Telegram.Timeout,
		},
		telegram.Dependencies{
			Auth:     authService,
			Commands: examinationCommandHandler,
			Queries:  examinationQueryHandler,
			Logger:   logger,
		},
	)
	if err != nil {
		logger.Error("failed to create telegram bot", "error", err)
		os.Exit(1)
	}

	processingWorker := command.NewProcessingWorker(
		examinationRepository,
		processingJobRepository,
		speechClient,
		// llmClient,
		logger,
		command.ProcessingWorkerConfig{
			Workers: cfg.Processing.Workers,
			Timeout: time.Duration(cfg.Processing.Timeout) * time.Second,
		},
	)

	outboxWorker := query.NewOutboxWorker(
		outboxRepository,
		examinationReadRepository,
		logger,
	)

	go func() {
		if err := processingWorker.Run(ctx); err != nil &&
			!errors.Is(err, context.Canceled) {

			logger.Error(
				"processing worker stopped with error",
				"error", err,
			)

			stop()
		}
	}()

	go func() {
		if err := outboxWorker.Run(ctx); err != nil &&
			!errors.Is(err, context.Canceled) {

			logger.Error(
				"outbox worker stopped with error",
				"error", err,
			)

			stop()
		}
	}()

	logger.Info(
		"application started",
		"telegram", true,
		"postgres", true,
		"redis", true,
		"speech_provider", cfg.Speech.Provider,
		"processing_workers", cfg.Processing.Workers,
	)

	go func() {
		logger.Info("telegram bot started")

		telegramBot.Start()

		logger.Info("telegram bot stopped")
	}()

	<-ctx.Done()

	logger.Info("shutdown signal received")

	telegramBot.Stop()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := processingWorker.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"failed to shutdown processing worker",
			"error", err,
		)
	}

	if err := outboxWorker.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"failed to shutdown outbox worker",
			"error", err,
		)
	}

	logger.Info("application stopped")
}
