package main

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(cfg *Config) error {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return err
	}

	slog.Info("bot started", "username", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if !IsAllowed(update.Message.From.ID, cfg.AllowedUserIDs) {
				continue
			}
			msg := update.Message
			if msg.IsCommand() {
				switch msg.Command() {
				case "start":
					HandleStart(bot, msg)
				case "status":
					HandleStatus(bot, msg)
				case "reboot":
					HandleReboot(bot, msg)
				}
			}
		}
		if update.CallbackQuery != nil {
			if !IsAllowed(update.CallbackQuery.From.ID, cfg.AllowedUserIDs) {
				continue
			}
			HandleCallback(bot, update.CallbackQuery)
		}
	}
	return nil
}
