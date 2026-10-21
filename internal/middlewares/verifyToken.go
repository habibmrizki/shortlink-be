package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/utils"
	"github.com/habibmrizki/shorlink-be/pkg"
	"github.com/redis/go-redis/v9"
)

func VerifyToken(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse("Authorization token is required"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse("Invalid authorization header format"))
			return
		}

		tokenString := parts[1]

		// Check Redis token blacklist
		if utils.IsBlacklisted(c.Request.Context(), rdb, tokenString) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse("Token has been logged out, please log in again"))
			return
		}

		claims := &pkg.Claims{}
		if err := claims.VerifyToken(tokenString); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized: invalid or expired token"))
			return
		}

		c.Set("user_id", claims.UserId)
		c.Set("role", claims.Role)
		c.Set("token", tokenString)
		if claims.ExpiresAt != nil {
			c.Set("expires_at", claims.ExpiresAt.Time)
		}

		c.Next()
	}
}
