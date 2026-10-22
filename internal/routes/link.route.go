package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/habibmrizki/shorlink-be/internal/handlers"
	"github.com/habibmrizki/shorlink-be/internal/middlewares"
	"github.com/redis/go-redis/v9"
)

func LinkRoutes(api *gin.RouterGroup, linkHandler *handlers.LinkHandler, rdb *redis.Client) {
	links := api.Group("/links")
	links.Use(middlewares.VerifyToken(rdb))

	links.POST("", linkHandler.CreateLink)
	links.GET("", linkHandler.GetLinks)

}
