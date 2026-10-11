package main

import (
	"log"
	"room-mate-api/internal/config"
	"room-mate-api/internal/database"

	"github.com/gin-gonic/gin"
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
	_ = db

	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})
	router.Run()
}
