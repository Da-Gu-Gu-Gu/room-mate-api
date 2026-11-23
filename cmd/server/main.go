package main

import (
	"context"
	"log"
	"room-mate-api/internal/config"
	"room-mate-api/internal/database"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/router"
	"room-mate-api/internal/service"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.NewTaskService(repository.NewTaskRepository(db)).RunScheduler(ctx)

	apisRouter := router.SetupRouter(db, cfg)
	if err := apisRouter.Run(":" + cfg.ServerConfig.Port); err != nil {
		log.Fatal(err)
	}
}
