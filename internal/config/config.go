package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type DataBasseConfig struct {
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabasePort     string
}

type SERVERConfig struct {
	Port string
	Env  string
}

type Config struct {
	DataBasseConfig DataBasseConfig
	JWTSecret       string
	ServerConfig    SERVERConfig
}

func LoadEnv() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("load .env failed: %w", err)
	}

	return &Config{
		DataBasseConfig: DataBasseConfig{
			DatabaseName:     os.Getenv("DATABASE_NAME"),
			DatabaseUser:     os.Getenv("DATABASE_USER"),
			DatabasePassword: os.Getenv("DATABASE_PASSWORD"),
			DatabasePort:     os.Getenv("DATABASE_PORT"),
		},
		ServerConfig: SERVERConfig{
			Port: os.Getenv("SERVER_PORT"),
			Env:  os.Getenv("SERVER_ENV"),
		},
		JWTSecret: os.Getenv("JWT_SECRET"),
	}, nil
}
