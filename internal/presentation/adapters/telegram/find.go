package telegram

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

// handleFind выполняет поиск только по обследованиям текущего Telegram-пользователя.
func (b *Bot) handleFind(c telebot.Context) error {
	keyword := strings.TrimSpace(c.Message().Payload)
	if keyword == "" {
		_, err := b.bot.Send(c.Sender(), "Использование:\n/find <текст>")
		return err
	}
	items, err := b.queries.Find(context.Background(), ports.FindExaminationsQuery{DoctorID: c.Sender().ID, Keyword: keyword})
	if err != nil {
		b.logger.Error("failed to find examinations", "error", err)
		_, sendErr := b.bot.Send(c.Sender(), "Не удалось выполнить поиск.")
		return sendErr
	}
	if len(items) == 0 {
		_, err = b.bot.Send(c.Sender(), "По вашему запросу ничего не найдено.")
		return err
	}
	var message strings.Builder
	for _, item := range items {
		fmt.Fprintf(&message, "ID: %s\nСтатус: %s\nВыжимка: %s\n\n", item.ID, item.Status, item.Summary)
	}
	_, err = b.bot.Send(c.Sender(), message.String())
	return err
}
