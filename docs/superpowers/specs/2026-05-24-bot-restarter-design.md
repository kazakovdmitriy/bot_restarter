# Bot Restarter — Design Spec

## Overview
TG-бот на Go, который перезагружает VDS, на котором запущен, командой `sudo reboot`. Доступ ограничен списком разрешённых user ID.

## Features
- `/start` — приветственное сообщение
- `/reboot` — inline-клавиатура с подтверждением [Да, перезагрузить] [Отмена]
- `/status` — вывод `uptime` и нагрузки

## Tech Stack
- Go 1.21+
- `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- `log/slog` для логирования
- config.json для настроек
- systemd-сервис + Makefile для сборки и деплоя

## File Structure
```
├── main.go              # Точка входа: читает конфиг, запускает бота
├── config.go            # Загрузка config.json
├── bot.go               # Инициализация бота, регистрация хендлеров
├── handlers.go          # Обработчики команд /start, /reboot, /status
├── auth.go              # Проверка allowed_user_ids
├── system.go            # Выполнение системных команд (uptime, reboot)
├── config.example.json  # Пример конфига
├── Makefile             # build, deploy, install-service
└── bot-restarter.service # systemd unit
```

## Config (config.json)
```json
{
  "telegram_token": "YOUR_BOT_TOKEN",
  "allowed_user_ids": [123456789, 987654321]
}
```

## Architecture
- Polling-based (long polling), без вебхуков
- Каждый апдейт проходит auth-проверку по `message.From.ID`
- Команды роутятся через switch/case
- Системные вызовы через `os/exec`

## Security
- `sudo reboot` требует настройки sudoers: `botuser ALL=(ALL) NOPASSWD: /sbin/reboot`
- Бот игнорирует сообщения от неавторизованных пользователей

## Deployment
- systemd unit с `Restart=always`, `User=botuser`
- Makefile: `make build` (linux/amd64), `make deploy` (scp + restart service), `make status`
