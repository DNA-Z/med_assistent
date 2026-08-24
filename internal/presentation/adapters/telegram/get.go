package telegram

import (
	"context"
	"fmt"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleGet(c telebot.Context) error {
	id, err := parseCommandUUID(c.Message())
	if err != nil {
		_, sendErr := b.bot.Send(
			c.Sender(),
			"Использование:\n/get <id>",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	result, err := b.queries.Get(
		context.Background(),
		ports.GetExaminationQuery{
			DoctorID:      c.Sender().ID,
			ExaminationID: id,
		},
	)
	if err != nil {
		b.logger.Error(
			"не удалось получить обследование",
			"telegram_user_id", c.Sender().ID,
			"examination_id", id,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось получить обследование.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	message := fmt.Sprintf(
		"Обследование: %s\n"+
			"Создано: %s\n"+
			"Статус: %s\n\n"+
			"Транскрипция:\n%s",
		result.ID,
		result.ExaminationDate.Format("02.01.2006 15:04"),
		result.Status,
		result.Transcript,
	)

	if result.Summary != "" {
		message += "\n\nВыжимка:\n" + result.Summary
	}

	_, err = b.bot.Send(c.Sender(), message)

	return err
}
