# Bot Restarter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** TG-бот на Go для перезагрузки VDS через `sudo reboot` с авторизацией по user ID.

**Architecture:** Polling-бот на `go-telegram-bot-api/v5`. Конфиг из JSON. systemd-сервис + Makefile для деплоя.

**Tech Stack:** Go 1.21+, go-telegram-bot-api/v5, log/slog, os/exec

---

### Task 1: Initialize Go module and dependencies

**Files:**
- Create: `go.mod`

- [ ] **Step 1: Init Go module**

```bash
go mod init bot_restarter
```

- [ ] **Step 2: Add telegram-bot-api dependency**

```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

- [ ] **Step 3: Verify module compiles**

```bash
go build ./...
```
Expected: no errors (just unused imports warning is fine at this stage)

---

### Task 2: Config loading

**Files:**
- Create: `config.go`
- Create: `config.example.json`

- [ ] **Step 1: Write config.go**

```go
package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	TelegramToken  string `json:"telegram_token"`
	AllowedUserIDs []int64 `json:"allowed_user_ids"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("telegram_token is required")
	}
	return &cfg, nil
}
```

- [ ] **Step 2: Write config.example.json**

```json
{
  "telegram_token": "123456:ABC-DEF1234ghikl-zyx57W2v1u123ew11",
  "allowed_user_ids": [123456789]
}
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./...
```

---

### Task 3: Auth check

**Files:**
- Create: `auth.go`

- [ ] **Step 1: Write auth.go**

```go
package main

import "slices"

func IsAllowed(userID int64, allowedIDs []int64) bool {
	return slices.Contains(allowedIDs, userID)
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./...
```

---

### Task 4: System commands

**Files:**
- Create: `system.go`

- [ ] **Step 1: Write system.go**

```go
package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetUptime() (string, error) {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return "", fmt.Errorf("uptime: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func Reboot() error {
	return exec.Command("sudo", "reboot").Run()
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./...
```

---

### Task 5: Command handlers

**Files:**
- Create: `handlers.go`

- [ ] **Step 1: Write handlers.go**

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./...
```

---

### Task 6: Bot initialization and main

**Files:**
- Create: `bot.go`
- Create: `main.go`

- [ ] **Step 1: Write bot.go**

```go
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
```

- [ ] **Step 2: Write main.go**

```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := LoadConfig("config.json")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := StartBot(cfg); err != nil {
		slog.Error("bot failed", "error", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Build**

```bash
go build -o bot-restarter .
```
Expected: binary `bot-restarter` created

---

### Task 7: Add missing import (config.go)

**Files:**
- Modify: `config.go`

Fix config.go — add `fmt` import:

```go
import (
	"encoding/json"
	"fmt"
	"os"
)
```

---

### Task 8: Makefile and systemd unit

**Files:**
- Create: `Makefile`
- Create: `bot-restarter.service`

- [ ] **Step 1: Write Makefile**

```makefile
BINARY := bot-restarter
REMOTE_HOST ?= your-server
REMOTE_USER ?= root
REMOTE_DIR ?= /opt/bot-restarter

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY) .

.PHONY: deploy
deploy: build
	scp $(BINARY) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	scp config.json $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "systemctl restart bot-restarter"

.PHONY: install-service
install-service:
	scp bot-restarter.service $(REMOTE_USER)@$(REMOTE_HOST):/etc/systemd/system/
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "systemctl daemon-reload && systemctl enable bot-restarter"

.PHONY: status
status:
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "systemctl status bot-restarter"

.PHONY: logs
logs:
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "journalctl -u bot-restarter -f"
```

- [ ] **Step 2: Write bot-restarter.service**

```ini
[Unit]
Description=Bot Restarter
After=network.target

[Service]
Type=simple
User=botuser
WorkingDirectory=/opt/bot-restarter
ExecStart=/opt/bot-restarter/bot-restarter
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

---

### Task 9: Final build and verification

- [ ] **Step 1: Build for linux/amd64**

```bash
make build
```

- [ ] **Step 2: Check binary exists**

```bash
ls -la bot-restarter
```

- [ ] **Step 3: Verify Go vet passes**

```bash
go vet ./...
```
