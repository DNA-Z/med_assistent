package telegram

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleList(c telebot.Context) error {
	result, err := b.queries.List(
		context.Background(),
		ports.ListExaminationsQuery{
			DoctorTelegramID: c.Sender().ID,
		},
	)
	if err != nil {
		b.logger.Error(
			"failed to list examinations",
			"telegram_user_id", c.Sender().ID,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось получить список обследований.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	if len(result) == 0 {
		_, err := b.bot.Send(
			c.Sender(),
			"У вас пока нет сохранённых обследований.",
		)

		return err
	}

	var message strings.Builder

	message.WriteString("Ваши обследования:\n\n")

	for _, examination := range result {
		fmt.Fprintf(
			&message,
			"ID: %s\n"+
				"Создано: %s\n"+
				"Статус: %s\n",
			examination.ID,
			examination.CreatedAt.Format("02.01.2006 15:04"),
			examination.Status,
		)

		if examination.Summary != "" {
			fmt.Fprintf(
				&message,
				"Выжимка: %s\n",
				examination.Summary,
			)
		}

		message.WriteString("\n")
	}

	_, err = b.bot.Send(c.Sender(), message.String())

	return err
}
