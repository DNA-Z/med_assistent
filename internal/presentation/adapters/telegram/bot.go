package telegram

import (
	"log/slog"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type Bot struct {
	bot      *telebot.Bot
	auth     ports.AuthService
	commands ports.ExaminationCommandHandler
	queries  ports.ExaminationQueryHandler
	logger   *slog.Logger
}

type Config struct {
	Token   string
	Timeout int
}

type Dependencies struct {
	Auth     ports.AuthService
	Commands ports.ExaminationCommandHandler
	Queries  ports.ExaminationQueryHandler
	Logger   *slog.Logger
}

func New(
	cfg Config,
	deps Dependencies,
) (*Bot, error) {

	b, err := telebot.NewBot(telebot.Settings{
		Token: cfg.Token,
		Poller: &telebot.LongPoller{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	})

	if err != nil {
		return nil, err
	}

	result := &Bot{
		bot:      b,
		auth:     deps.Auth,
		commands: deps.Commands,
		queries:  deps.Queries,
		logger:   deps.Logger,
	}

	result.registerHandlers()

	return result, nil
}
