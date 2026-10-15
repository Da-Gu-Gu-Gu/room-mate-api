package main

import (
	"log"
	"room-mate-api/internal/config"
	"room-mate-api/internal/database"
	"room-mate-api/internal/router"
)

func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		panic(err)
	}
	db, err := database.Load(cfg)
	if err != nil {
		log.Fatal("got error while connecting to database")
	}

	apisRouter := router.SetupRouter(db)
	if err := apisRouter.Run(":" + cfg.ServerConfig.Port); err != nil {
		log.Fatal(err)
	}
}
