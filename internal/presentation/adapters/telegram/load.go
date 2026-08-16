package telegram

import (
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
	)
}

func (b *Bot) handleAudio(c telebot.Context) error {
	audio := c.Message().Audio

	return b.processTelegramFile(
		c,
		audio.FileID,
		audio.FileName,
	)
}

func (b *Bot) processTelegramFile(
	c telebot.Context,
	fileID string,
	fileName string,
) error {
	file := &telebot.File{
		FileID: fileID,
	}

	if err := b.bot.Download(&file.File, fileName); err != nil {
		b.logger.Error(
			"failed to download telegram file",
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

	data, err := readFile(fileName)
	if err != nil {
		b.logger.Error(
			"failed to read downloaded file",
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
			DoctorID: c.Sender().ID,
			File:     data,
			FileName: fileName,
		},
	)
	if err != nil {
		b.logger.Error(
			"failed to load examination",
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
