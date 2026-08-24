package telegram

import (
	"strings"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"
)

// parseCommandUUID извлекает идентификатор из аргумента Telegram-команды.
// TrimSpace защищает разбор от пробелов и переводов строки при копировании команды.
func parseCommandUUID(message *telebot.Message) (uuid.UUID, error) {
	if message == nil {
		return uuid.Parse("")
	}

	return uuid.Parse(strings.TrimSpace(message.Payload))
}
