package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/habibmrizki/shorlink-be/internal/dto"
	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/repositories"
	"github.com/habibmrizki/shorlink-be/internal/utils"
	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidURL       = utils.ErrInvalidURL
	ErrSlugTaken        = utils.ErrSlugTaken
	ErrSlugInvalidLen   = utils.ErrSlugInvalidLen
	ErrSlugInvalidChars = utils.ErrSlugInvalidChars
	ErrSlugReserved     = utils.ErrSlugReserved
	ErrLinkNotFound     = utils.ErrLinkNotFound
)

type LinkService interface {
	CreateLink(ctx context.Context, userID int, req *dto.CreateLinkRequest, baseURL string) (*dto.LinkResponse, error)
	GetLinks(ctx context.Context, userID int, baseURL string) ([]dto.LinkResponse, error)
	GetLinksWithPagination(ctx context.Context, userID int, page, limit int, showDeleted bool, baseURL string) (*dto.PaginatedLinkResponse, error)
	DeleteLink(ctx context.Context, linkID, userID int) error
	GetOriginalURL(ctx context.Context, slug string) (string, error)
	CheckSlugAvailability(ctx context.Context, slug string) (bool, string, error)
}

type linkService struct {
	linkRepo repositories.LinkRepository
	rdb      *redis.Client
}

func NewLinkService(linkRepo repositories.LinkRepository, rdb *redis.Client) LinkService {
	return &linkService{
		linkRepo: linkRepo,
		rdb:      rdb,
	}
}

func (s *linkService) CreateLink(ctx context.Context, userID int, req *dto.CreateLinkRequest, baseURL string) (*dto.LinkResponse, error) {
	// Validasi Original URL
	parsedURL, err := url.ParseRequestURI(req.OriginalURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, ErrInvalidURL
	}

	// Tentukan Slug (BEebas Custom atau Auto-generated)
	targetSlug := strings.TrimSpace(req.Slug)
	if targetSlug == "" {
		targetSlug = strings.TrimSpace(req.CustomSlug)
	}

	var slug string
	if targetSlug != "" {
		if len(targetSlug) < 3 || len(targetSlug) > 50 {
			return nil, ErrSlugInvalidLen
		}
		if !utils.IsValidSlugFormat(targetSlug) {
			return nil, ErrSlugInvalidChars
		}
		if utils.IsReservedSlug(targetSlug) {
			return nil, ErrSlugReserved
		}

		exists, err := s.linkRepo.IsSlugExists(ctx, targetSlug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrSlugTaken
		}
		slug = targetSlug
	} else {
		generated, err := s.generateUniqueSlug(ctx)
		if err != nil {
			return nil, err
		}
		slug = generated
	}

	// Simpan ke DB Transaction
	link := &models.Link{
		UserId:      userID,
		OriginalURL: req.OriginalURL,
		Slug:        slug,
	}

	created, err := s.linkRepo.CreateLink(ctx, link)
	if err != nil {
		return nil, err
	}

	// Simpan ke Redis cache 24 jam
	_ = utils.SetSlugCache(ctx, s.rdb, created.Slug, created.OriginalURL, 24*time.Hour)

	cleanBaseURL := strings.TrimRight(baseURL, "/")
	shortURL := fmt.Sprintf("%s/%s", cleanBaseURL, created.Slug)

	return &dto.LinkResponse{
		ID:          created.Id,
		OriginalURL: created.OriginalURL,
		Slug:        created.Slug,
		ShortURL:    shortURL,
		ClickCount:  created.ClickCount,
		CreatedAt:   created.CreatedAt,
		DeletedAt:   created.DeletedAt,
	}, nil
}
