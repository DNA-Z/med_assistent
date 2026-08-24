package telegram

import (
	"context"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleDelete(c telebot.Context) error {
	id, err := uuid.Parse(c.Message().Payload)
	if err != nil {
		_, sendErr := b.bot.Send(
			c.Sender(),
			"Использование:\n/delete <id>",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	err = b.commands.Delete(
		context.Background(),
		ports.DeleteExaminationCommand{
			DoctorID:      c.Sender().ID,
			ExaminationID: id,
		},
	)
	if err != nil {
		b.logger.Error(
			"не удалось удалить обследование",
			"telegram_user_id", c.Sender().ID,
			"examination_id", id,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось удалить обследование.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	_, err = b.bot.Send(
		c.Sender(),
		"Обследование удалено.",
	)

	return err
}
