package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/habibmrizki/shorlink-be/internal/binder"
	"github.com/habibmrizki/shorlink-be/internal/dto"
	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/services"
	"github.com/habibmrizki/shorlink-be/internal/utils"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register creates a new user account.
// @Summary Register a user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 409 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := binder.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Validation error",
			Results: validationErrors,
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, utils.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, models.ErrorResponse("Email already registered"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("User registered successfully", user))
}

// Login authenticates a user and returns a JWT.
// @Summary Log in
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := binder.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Validation error",
			Results: validationErrors,
		})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, utils.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid email or password"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Login successful", result))
}
