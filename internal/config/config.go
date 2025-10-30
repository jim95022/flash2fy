package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/subosito/gotenv"
)

type (
	Server struct {
		Addr string
	}

	Database struct {
		URL string
	}

	Telegram struct {
		BotToken string
	}

	Config struct {
		Server   Server
		Database Database
		Telegram Telegram
	}
)

// Load reads configuration from environment variables, loading .env if present.
func Load() (*Config, error) {
	if err := loadEnvFile(); err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: Server{
			Addr: getEnv("SERVER_ADDR", ":8080"),
		},
		Database: Database{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/flash2fy?sslmode=disable"),
		},
		Telegram: Telegram{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		},
	}

	return cfg, nil
}

func loadEnvFile() error {
	err := gotenv.Load()
	if err == nil {
		return nil
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return nil
	}

	return fmt.Errorf("load .env file: %w", err)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
