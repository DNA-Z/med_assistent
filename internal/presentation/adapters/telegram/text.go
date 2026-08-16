package telegram

import (
	"gopkg.in/telebot.v3"
)

func (b *Bot) handleText(c telebot.Context) error {
	_, err := b.bot.Send(
		c.Sender(),
		"Неизвестная команда.\n\n"+
			"Используйте /start для просмотра доступных команд.",
	)

	return err
}
