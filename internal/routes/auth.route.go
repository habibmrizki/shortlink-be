package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/habibmrizki/shorlink-be/internal/handlers"
	"github.com/habibmrizki/shorlink-be/internal/repositories"
	"github.com/habibmrizki/shorlink-be/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func AuthRoutes(api *gin.RouterGroup, db *pgxpool.Pool, rdb *redis.Client) {
	authRepo := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepo, rdb)
	authHandler := handlers.NewAuthHandler(authService)

	api.POST("/register", authHandler.Register)
}
