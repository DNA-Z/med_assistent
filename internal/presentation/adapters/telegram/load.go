package telegram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"gopkg.in/telebot.v3"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func (b *Bot) handleLoad(c telebot.Context) error {
	if c.Message() == nil {
		return nil
	}

	_, err := b.bot.Send(
		c.Sender(),
		"Для загрузки встречи отправьте голосовое сообщение или аудиофайл.",
	)

	return err
}

func (b *Bot) handleVoice(c telebot.Context) error {
	return b.processTelegramFile(
		c,
		c.Message().Voice.FileID,
		"voice.ogg",
		c.Message().Voice.MIME,
	)
}

func (b *Bot) handleAudio(c telebot.Context) error {
	audio := c.Message().Audio

	return b.processTelegramFile(
		c,
		audio.FileID,
		audio.FileName,
		audio.MIME,
	)
}

func (b *Bot) processTelegramFile(
	c telebot.Context,
	fileID string,
	fileName string,
	contentType string,
) error {
	file := &telebot.File{
		FileID: fileID,
	}

	temp, err := os.CreateTemp("", "med-assistant-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	_ = temp.Close()
	defer os.Remove(tempName)
	if err := b.bot.Download(file, tempName); err != nil {
		b.logger.Error(
			"не удалось скачать файл из Telegram",
			"telegram_user_id", c.Sender().ID,
			"file_id", fileID,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось скачать файл.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	data, err := readFile(tempName)
	if err != nil {
		b.logger.Error(
			"не удалось прочитать скачанный файл",
			"telegram_user_id", c.Sender().ID,
			"file_name", fileName,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось прочитать файл.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	id, err := b.commands.Load(
		context.Background(),
		ports.LoadExaminationCommand{
			DoctorID:    c.Sender().ID,
			File:        io.NopCloser(bytes.NewReader(data)),
			FileName:    fileName,
			FileSize:    int64(len(data)),
			ContentType: contentType,
		},
	)
	if err != nil {
		b.logger.Error(
			"не удалось загрузить обследование",
			"telegram_user_id", c.Sender().ID,
			"file_name", fileName,
			"error", err,
		)

		_, sendErr := b.bot.Send(
			c.Sender(),
			"Не удалось создать обработку файла.",
		)

		if sendErr != nil {
			return sendErr
		}

		return nil
	}

	_, err = b.bot.Send(
		c.Sender(),
		fmt.Sprintf(
			"Файл принят.\n\n"+
				"Обследование: %s\n"+
				"Обработка выполняется в фоне.\n"+
				"Проверить статус: /status %s",
			id,
			id,
		),
	)

	return err
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
