package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DataBasseConfig struct {
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
}

type SERVERConfig struct {
	Port string
	Env  string
}

type Config struct {
	DataBasseConfig DataBasseConfig
	JWTSecret       string
	serverConfig    SERVERConfig
}

func LoadEnv() (*Config, error) {
	godotenv.Load()
	return &Config{
		DataBasseConfig: DataBasseConfig{
			DatabaseName:     os.Getenv("DATABASE-NAME"),
			DatabaseUser:     os.Getenv("DATABASE-USER"),
			DatabasePassword: os.Getenv("DATABASE-PASSWORD"),
		},
		serverConfig: SERVERConfig{
			Port: os.Getenv("SERVER-PORT"),
			Env:  os.Getenv("SERVER-ENV"),
		},
		JWTSecret: os.Getenv("JWT-SECRET"),
	}, nil
}
