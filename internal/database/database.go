package database

import (
	"fmt"
	"room-mate-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Load(cfg *config.Config) (*gorm.DB, error) {
	// Initialize the database connection using the provided configuration
	// For example, you can use GORM to connect to a PostgreSQL database
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable",
		cfg.DataBasseConfig.DatabaseUser,
		cfg.DataBasseConfig.DatabasePassword,
		cfg.DataBasseConfig.DatabaseName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
