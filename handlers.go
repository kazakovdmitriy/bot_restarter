package main

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	callbackRebootConfirm = "reboot_confirm"
	callbackRebootCancel  = "reboot_cancel"
)

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "Привет! Я бот для управления сервером.\n"
	text += "/status — состояние сервера\n"
	text += "/reboot — перезагрузить сервер"
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if _, err := bot.Send(reply); err != nil {
		slog.Error("send start", "error", err)
	}
}

func HandleStatus(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	uptime, err := GetUptime()
	if err != nil {
		slog.Error("get uptime", "error", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка получения статуса")
		bot.Send(reply)
		return
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("```\n%s\n```", uptime))
	reply.ParseMode = "MarkdownV2"
	bot.Send(reply)
}

func HandleReboot(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "Вы уверены, что хотите перезагрузить сервер?")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, перезагрузить", callbackRebootConfirm),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", callbackRebootCancel),
		),
	)
	if _, err := bot.Send(reply); err != nil {
		slog.Error("send reboot prompt", "error", err)
	}
}

func HandleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(cb.ID, "")
	bot.Send(callback)

	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "")
	edit.ReplyMarkup = nil

	switch cb.Data {
	case callbackRebootConfirm:
		edit.Text = "Перезагружаю сервер..."
		bot.Send(edit)
		slog.Info("reboot initiated", "user_id", cb.From.ID, "username", cb.From.UserName)
		if err := Reboot(); err != nil {
			slog.Error("reboot failed", "error", err)
		}
	case callbackRebootCancel:
		edit.Text = "Перезагрузка отменена"
		bot.Send(edit)
	}
}
