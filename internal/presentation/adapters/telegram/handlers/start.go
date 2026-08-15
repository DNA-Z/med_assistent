package handlers

import (
	"context"
	"fmt"
	"logger"

	"gopkg.in/telebot.v3"
)

func (b *Bot) handleStart(m *telebot.Message) {

	ctx := context.Background()

	err := b.auth.Start(ctx, m.Sender.ID)
	if err != nil {
		b.logger.Error(
			"failed to authenticate telegram user",
			"telegram_user_id", m.Sender.ID,
			"error", err,
		)

		_ = b.bot.Send(
			m.Sender,
			"Не удалось зарегистрировать пользователя.",
		)

		return
	}

	_, _ = b.bot.Send(
		m.Sender,
		"Вы успешно зарегистрированы.\n\n"+
			"Доступные команды:\n"+
			"/load — загрузить аудиозапись\n"+
			"/list — список обследований\n"+
			"/status <id> — статус обработки\n"+
			"/get <id> — получить транскрипцию\n"+
			"/find <текст> — поиск\n"+
			"/chat <вопрос> — задать вопрос\n"+
			"/retry <id> — повторить обработку",
	)
}
