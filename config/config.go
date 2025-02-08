package config

import (
	"os"
)

type Config struct {
	DatabaseURL   string `json:"database_url"`
	TelegramToken string `json:"telegram_token"`
}

func LoadConfig() (Config, error) {
	config := Config{DatabaseURL: os.Getenv("DATABASE_URL"), TelegramToken: os.Getenv("TELEGRAM_TOKEN")}
	return config, nil
}
