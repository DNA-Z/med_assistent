package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DNA-Z/med_assistent/internal/infrastructure/config"
	"golang.org/x/sync/errgroup"
)

func TestValidateConfig(t *testing.T) {
	cfg := config.NewConfig()
	cfg.DBConnectionString = "postgres://test"
	cfg.Redis.Address = "localhost:6379"
	cfg.Telegram.Token = "token"
	cfg.ObjectStorage.Endpoint = "localhost:9000"
	cfg.ObjectStorage.AccessKey = "access"
	cfg.ObjectStorage.SecretKey = "secret"
	cfg.ObjectStorage.Bucket = "audio"
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	cfg.Telegram.Token = ""
	if err := validateConfig(cfg); err == nil || !strings.Contains(err.Error(), "TELEGRAM_TOKEN") {
		t.Fatalf("err=%v", err)
	}
}

func TestWaitForGroupTimesOutDuringShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	group, _ := errgroup.WithContext(ctx)
	release := make(chan struct{})
	group.Go(func() error { <-release; return nil })
	cancel()
	err := waitForGroup(ctx, group, time.Millisecond)
	close(release)
	if err == nil {
		t.Fatal("timeout error expected")
	}
}

func TestWaitForGroupWaitsForAllWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { <-groupCtx.Done(); return nil })
	cancel()
	if err := waitForGroup(ctx, group, time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestClientsRequireOpenAIAPIKey(t *testing.T) {
	cfg := config.NewConfig()
	cfg.LLM.Provider = "openai"

	if _, _, err := clients(cfg); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("err=%v", err)
	}
}
