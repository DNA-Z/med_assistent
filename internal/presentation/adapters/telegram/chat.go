package telegram

import (
	"context"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

// handleChat передаёт вопрос пользователя query-обработчику и отправляет
// полученный от LLM ответ обратно в Telegram.
func (b *Bot) handleChat(c telebot.Context) error {
	question := strings.TrimSpace(c.Message().Payload)
	if question == "" {
		_, err := b.bot.Send(c.Sender(), "Использование:\n/chat <вопрос>")
		return err
	}
	answer, err := b.queries.Chat(context.Background(), ports.ChatQuery{DoctorID: c.Sender().ID, Question: question})
	if err != nil {
		b.logger.Error("не удалось получить ответ чата", "error", err)
		_, sendErr := b.bot.Send(c.Sender(), "Не удалось получить ответ.")
		return sendErr
	}
	_, err = b.bot.Send(c.Sender(), answer)
	return err
}
