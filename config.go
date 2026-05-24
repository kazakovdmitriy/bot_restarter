package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	TelegramToken  string  `json:"telegram_token"`
	AllowedUserIDs []int64 `json:"allowed_user_ids"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("telegram_token is required")
	}
	return &cfg, nil
}
