package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/habibmrizki/shorlink-be/docs"
	"github.com/habibmrizki/shorlink-be/internal/handlers"
	"github.com/habibmrizki/shorlink-be/internal/middlewares"
	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/repositories"
	"github.com/habibmrizki/shorlink-be/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	router := gin.Default()

	router.Use(middlewares.CORSMiddleware)

	linkRepo := repositories.NewLinkRepository(db)
	linkService := services.NewLinkService(linkRepo, rdb)
	linkHandler := handlers.NewLinkHandler(linkService)

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.SuccessResponse("Server is healthy", nil))
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")

	AuthRoutes(api, db, rdb)
	LinkRoutes(api, linkHandler, rdb)

	router.GET("/:slug", linkHandler.Redirect)
	router.HEAD("/:slug", linkHandler.Redirect)

	return router
}
