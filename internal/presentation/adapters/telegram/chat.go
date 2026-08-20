package telegram

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"gopkg.in/telebot.v3"
	"strings"
)

func (b *Bot) handleChat(c telebot.Context) error {
	question := strings.TrimSpace(c.Message().Payload)
	if question == "" {
		_, err := b.bot.Send(c.Sender(), "Использование:\n/chat <вопрос>")
		return err
	}
	answer, err := b.queries.Chat(context.Background(), ports.ChatQuery{DoctorID: c.Sender().ID, Question: question})
	if err != nil {
		b.logger.Error("failed to chat", "error", err)
		_, sendErr := b.bot.Send(c.Sender(), "Не удалось получить ответ.")
		return sendErr
	}
	_, err = b.bot.Send(c.Sender(), answer)
	return err
}

//
//import (
//	"context"
//	"strings"
//
//	"gopkg.in/telebot.v3"
//
//	"github.com/DNA-Z/med_assistent/internal/application/ports"
//)
//
//func (b *Bot) handleChat(c telebot.Context) error {
//	question := strings.TrimSpace(c.Message().Payload)
//
//	if question == "" {
//		_, err := b.bot.Send(
//			c.Sender(),
//			"Использование:\n/chat <вопрос>",
//		)
//
//		return err
//	}
//
//	result, err := b.queries.Chat(
//		context.Background(),
//		ports.ChatQuery{
//			DoctorTelegramID: c.Sender().ID,
//			Question:         question,
//		},
//	)
//	if err != nil {
//		b.logger.Error(
//			"failed to chat with llm",
//			"telegram_user_id", c.Sender().ID,
//			"error", err,
//		)
//
//		_, sendErr := b.bot.Send(
//			c.Sender(),
//			"Не удалось получить ответ.",
//		)
//
//		if sendErr != nil {
//			return sendErr
//		}
//
//		return nil
//	}
//
//	_, err = b.bot.Send(
//		c.Sender(),
//		result.Answer,
//	)
//
//	return err
//}
