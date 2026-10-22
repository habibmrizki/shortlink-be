package main

import (
	"log"

	"github.com/habibmrizki/shorlink-be/internal/configs"
	"github.com/habibmrizki/shorlink-be/internal/routes"
	"github.com/joho/godotenv"
)

// @title Shortlink API
// @version 1.0
// @description API for user authentication, short link management, and URL redirection.
// @host localhost:2901
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	// Connect PostgreSQL
	db, err := configs.InitDB()
	if err != nil {
		log.Println("FAILED TO CONNECT DB:", err.Error())
		return
	}
	defer db.Close()

	if err := configs.PingDB(db); err != nil {
		log.Println("PING TO DB FAILED:", err.Error())
		return
	}
	log.Println("DB CONNECTED")

	// Connect Redis
	rdb := configs.InitRedis()
	if err := configs.PingRedis(rdb); err != nil {
		log.Println("WARNING: REDIS CONNECTION FAILED:", err.Error())
	} else {
		log.Println("REDIS CONNECTED")
	}
	defer rdb.Close()

	// Initialize Router
	r := routes.InitRouter(db, rdb)

	r.Run(":2901")
}
