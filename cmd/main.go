package main

import (
	"log"

	_ "github.com/habibmrizki/shorlink-be/docs"
	"github.com/habibmrizki/shorlink-be/internal/configs"
	"github.com/habibmrizki/shorlink-be/internal/routes"
	"github.com/joho/godotenv"
)

// @title Shortlink API
// @version 1.0
// @description User authentication API for the Shortlink service.
// @host localhost:2901
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	db, err := configs.InitDB()
	if err != nil {
		log.Println("FAILED TO CONNECT DB")
		return
	}

	defer db.Close()

	err = configs.PingDB(db)
	if err != nil {
		log.Println("PING TO DB FAILED", err.Error())
		return
	}

	log.Println("DB CONNECTED")

	r := routes.InitRouter(db)

	r.Run(":2901")
}
