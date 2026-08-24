package telegram

import (
	"testing"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"
)

func TestParseCommandUUIDTrimsWhitespace(t *testing.T) {
	t.Parallel()

	want := uuid.MustParse("40f13b14-2f9d-4a27-83bd-c641b413fb67")
	message := &telebot.Message{Payload: "  \n" + want.String() + "\r\n "}

	got, err := parseCommandUUID(message)
	if err != nil {
		t.Fatalf("разобрать идентификатор: %v", err)
	}
	if got != want {
		t.Fatalf("получен идентификатор %s, ожидался %s", got, want)
	}
}

func TestParseCommandUUIDRejectsMissingArgument(t *testing.T) {
	t.Parallel()

	if _, err := parseCommandUUID(&telebot.Message{}); err == nil {
		t.Fatal("ожидалась ошибка для пустого аргумента команды")
	}
}
