package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	Env       string
	JWTSecret string
	Database  struct {
		DSN string //Data Source Name to reduce creation of more fields for username/password
	}
	Redis struct {
		Address  string
		Password string
	}
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	cfg := &Config{}

	cfg.Env = os.Getenv("ENV")
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	cfg.Database.DSN = os.Getenv("DATABASE_DSN")
	if cfg.Database.DSN == "" {
		return nil, errors.New("error: DATABASE_DSN environment variable is required")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, errors.New("error: JWT_SECRET environment variable is required")
	}

	cfg.Redis.Address = os.Getenv("REDIS_ADDRESS")
	if cfg.Redis.Address == "" {
		return nil, errors.New("error: REDIS_ADDRESS environment variable is required")
	}
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	return cfg, nil
}
