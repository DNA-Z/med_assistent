package telegram

import (
	"context"

	"gopkg.in/telebot.v3"
)

func (b *Bot) handleStart(c telebot.Context) error {
	ctx := context.Background()

	err := b.auth.Start(ctx, c.Sender().ID)
	if err != nil {
		b.logger.Error(
			"failed to authenticate telegram user",
			"telegram_user_id", c.Sender().ID,
			"error", err,
		)

		if _, err := b.bot.Send(
			c.Sender(),
			"Не удалось зарегистрировать пользователя.",
		); err != nil {
			return err
		}

		return nil
	}

	_, err = b.bot.Send(
		c.Sender(),
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

	return err
}
