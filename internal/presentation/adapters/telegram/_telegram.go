package telegram

import "gopkg.in/telebot.v3"

func (b *Bot) registerHandlers() {
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/load", b.handleLoad)
	b.bot.Handle("/list", b.handleList)
	b.bot.Handle("/status", b.handleStatus)
	b.bot.Handle("/get", b.handleGet)
	b.bot.Handle("/find", b.handleFind)
	b.bot.Handle("/chat", b.handleChat)
	b.bot.Handle("/retry", b.handleRetry)
	b.bot.Handle("/delete", b.handleDelete)
	b.bot.Handle(telebot.OnVoice, b.handleVoice)
	b.bot.Handle(telebot.OnAudio, b.handleAudio)
	b.bot.Handle(telebot.OnText, b.handleText)
}
