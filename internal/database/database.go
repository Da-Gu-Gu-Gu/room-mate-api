package database

import (
	"fmt"
	"room-mate-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Load(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DataBasseConfig.DatabaseUser,
		cfg.DataBasseConfig.DatabasePassword,
		cfg.DataBasseConfig.DatabaseName,
		cfg.DataBasseConfig.DatabasePort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
