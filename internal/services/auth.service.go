package services

import (
	"context"
	"time"

	"github.com/habibmrizki/shorlink-be/internal/dto"
	"github.com/habibmrizki/shorlink-be/internal/repositories"
	"github.com/habibmrizki/shorlink-be/internal/utils"
	"github.com/habibmrizki/shorlink-be/pkg"
	"github.com/redis/go-redis/v9"
)

var (
	ErrEmailAlreadyExists = utils.ErrEmailAlreadyExists
	ErrInvalidCredentials = utils.ErrInvalidCredentials
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResult, error)
	Logout(ctx context.Context, token string, expiresAt time.Time) error
}

type authService struct {
	authRepo repositories.AuthRepository
	rdb      *redis.Client
}

func NewAuthService(authRepo repositories.AuthRepository, rdb *redis.Client) AuthService {
	return &authService{
		authRepo: authRepo,
		rdb:      rdb,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	exists, err := s.authRepo.IsEmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	hashCfg := pkg.NewHashConfig()
	passwordHash, err := hashCfg.GenHash(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.authRepo.CreateUser(ctx, req.Email, passwordHash)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.Id,
		Email:     user.Email,
		CreatedAt: &user.CreatedAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResult, error) {
	user, err := s.authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	hashCfg := pkg.NewHashConfig()
	isValid, err := hashCfg.CompareHashAndPassword(req.Password, user.PasswordHash)
	if err != nil || !isValid {
		return nil, ErrInvalidCredentials
	}

	claims := pkg.NewJWTClaims(user.Id, "user")
	token, err := claims.GenToken()
	if err != nil {
		return nil, err
	}

	return &dto.LoginResult{
		Token: token,
		User: dto.UserResponse{
			ID:        user.Id,
			Email:     user.Email,
			CreatedAt: &user.CreatedAt,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, token string, expiresAt time.Time) error {
	return utils.AddToBlacklist(ctx, s.rdb, token, expiresAt)
}
