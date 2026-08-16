package telegram

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleFind(c telebot.Context) error {
	keyword := strings.TrimSpace(c.Message().Payload)

	if keyword == "" {
		_, err := b.bot.Send(
			c.Sender(),
			"Использование:\n/find <текст>",
		)

		return err
	}

	result, err := b.queries.Find(
		context.Background(),
		ports.FindExaminationsQuery{
			DoctorTelegramID: c.Sender().ID,
			Keyword:          keyword,
		},
	)
	if err != nil {
		b.logger.Error(
			"failed to find examinations",
			"telegram_user_id", c.Sender().ID,
			"keyword", keyword,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось выполнить поиск.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	if len(result) == 0 {
		_, err := b.bot.Send(
			c.Sender(),
			"По вашему запросу ничего не найдено.",
		)

		return err
	}

	var message strings.Builder

	fmt.Fprintf(
		&message,
		"Результаты поиска: %q\n\n",
		keyword,
	)

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

		if examination.Snippet != "" {
			fmt.Fprintf(
				&message,
				"Фрагмент: %s\n",
				examination.Snippet,
			)
		}

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
