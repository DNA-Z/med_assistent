package telegram

import (
	"context"
	"fmt"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleStatus(c telebot.Context) error {
	id, err := parseCommandUUID(c.Message())
	if err != nil {
		_, sendErr := b.bot.Send(
			c.Sender(),
			"Использование:\n/status <id>",
		)

		return sendErr
	}

	result, err := b.queries.Status(
		context.Background(),
		ports.GetExaminationStatusQuery{
			DoctorID:      c.Sender().ID,
			ExaminationID: id,
		},
	)

	if err != nil {
		b.logger.Error(
			"не удалось получить статус обследования",
			"telegram_user_id", c.Sender().ID,
			"examination_id", id,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось получить статус обследования.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	message := fmt.Sprintf(
		"Обследование: %s\n"+
			"Статус: %s\n"+
			"Создано: %s\n"+
			"Изменено: %s",
		result.ID,
		result.Status,
		result.CreatedAt.Format("02.01.2006 15:04"),
		result.UpdatedAt.Format("02.01.2006 15:04"),
	)

	if result.Error != "" {
		message += "\nОшибка: " + result.Error
	}

	_, err = b.bot.Send(c.Sender(), message)

	return err
}
