package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/habibmrizki/shorlink-be/internal/binder"
	"github.com/habibmrizki/shorlink-be/internal/dto"
	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/services"
	"github.com/habibmrizki/shorlink-be/internal/utils"
)

type LinkHandler struct {
	linkService services.LinkService
}

func NewLinkHandler(linkService services.LinkService) *LinkHandler {
	return &LinkHandler{linkService: linkService}
}

// CreateLink creates a short link for the authenticated user.
// @Summary Create a short link
// @Tags links
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateLinkRequest true "Original URL and optional custom slug"
// @Success 201 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/links [post]
func (h *LinkHandler) CreateLink(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized"))
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid user identification"))
		return
	}

	var req dto.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := binder.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Validation error",
			Results: validationErrors,
		})
		return
	}

	baseURL := getBaseURL(c)
	result, err := h.linkService.CreateLink(c.Request.Context(), userID, &req, baseURL)
	if err != nil {
		if errors.Is(err, utils.ErrSlugTaken) {
			c.JSON(http.StatusConflict, models.ErrorResponse(err.Error()))
			return
		}
		if errors.Is(err, utils.ErrInvalidURL) ||
			errors.Is(err, utils.ErrSlugInvalidLen) ||
			errors.Is(err, utils.ErrSlugInvalidChars) ||
			errors.Is(err, utils.ErrSlugReserved) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Link created successfully", result))
}

// GetLinks lists short links created by the authenticated user with optional pagination and soft-delete listing.
// @Summary Get user's links
// @Tags links
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page limit"
// @Param show_deleted query bool false "Show soft-deleted links"
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/links [get]
func (h *LinkHandler) GetLinks(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized"))
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid user identification"))
		return
	}

	baseURL := getBaseURL(c)
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	showDeletedStr := c.Query("show_deleted")

	if pageStr != "" || limitStr != "" || showDeletedStr != "" {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(limitStr)
		if limit < 1 {
			limit = 10
		}
		showDeleted := showDeletedStr == "true" || showDeletedStr == "1"

		paginatedResult, err := h.linkService.GetLinksWithPagination(c.Request.Context(), userID, page, limit, showDeleted, baseURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusOK, models.SuccessResponse("Links retrieved successfully", paginatedResult))
		return
	}

	results, err := h.linkService.GetLinks(c.Request.Context(), userID, baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Links retrieved successfully", results))
}

func getBaseURL(c *gin.Context) string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL != "" {
		return baseURL
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}
