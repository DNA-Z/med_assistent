package telegram

import (
	"context"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleRetry(c telebot.Context) error {
	id, err := uuid.Parse(c.Message().Payload)
	if err != nil {
		_, sendErr := b.bot.Send(
			c.Sender(),
			"Использование:\n/retry <id>",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	err = b.commands.Retry(
		context.Background(),
		ports.RetryExaminationCommand{
			DoctorID:      c.Sender().ID,
			ExaminationID: id,
		},
	)
	if err != nil {
		b.logger.Error(
			"failed to retry examination",
			"telegram_user_id", c.Sender().ID,
			"examination_id", id,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось повторно запустить обработку.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	_, err = b.bot.Send(
		c.Sender(),
		"Обработка обследования запущена повторно.",
	)

	return err
}
