package services

import (
	"context"
	"errors"
	"fmt"
	"math"
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

func (s *linkService) GetLinks(ctx context.Context, userID int, baseURL string) ([]dto.LinkResponse, error) {
	links, err := s.linkRepo.GetLinksByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cleanBaseURL := strings.TrimRight(baseURL, "/")
	responses := make([]dto.LinkResponse, 0, len(links))
	for _, l := range links {
		responses = append(responses, dto.LinkResponse{
			ID:          l.Id,
			OriginalURL: l.OriginalURL,
			Slug:        l.Slug,
			ShortURL:    fmt.Sprintf("%s/%s", cleanBaseURL, l.Slug),
			ClickCount:  l.ClickCount,
			CreatedAt:   l.CreatedAt,
			DeletedAt:   l.DeletedAt,
		})
	}

	return responses, nil
}

func (s *linkService) GetLinksWithPagination(ctx context.Context, userID int, page, limit int, showDeleted bool, baseURL string) (*dto.PaginatedLinkResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	links, totalItems, err := s.linkRepo.GetLinksByUserIDWithPagination(ctx, userID, page, limit, showDeleted)
	if err != nil {
		return nil, err
	}

	cleanBaseURL := strings.TrimRight(baseURL, "/")
	data := make([]dto.LinkResponse, 0, len(links))
	for _, l := range links {
		data = append(data, dto.LinkResponse{
			ID:          l.Id,
			OriginalURL: l.OriginalURL,
			Slug:        l.Slug,
			ShortURL:    fmt.Sprintf("%s/%s", cleanBaseURL, l.Slug),
			ClickCount:  l.ClickCount,
			CreatedAt:   l.CreatedAt,
			DeletedAt:   l.DeletedAt,
		})
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &dto.PaginatedLinkResponse{
		Data: data,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *linkService) DeleteLink(ctx context.Context, linkID, userID int) error {
	link, err := s.linkRepo.GetLinkByIDAndUserID(ctx, linkID, userID)
	if err != nil {
		return err
	}
	if link == nil {
		return ErrLinkNotFound
	}

	if err := s.linkRepo.DeleteLink(ctx, linkID, userID); err != nil {
		return err
	}

	_ = utils.InvalidateSlugCache(ctx, s.rdb, link.Slug)
	return nil
}

func (s *linkService) GetOriginalURL(ctx context.Context, slug string) (string, error) {
	if utils.IsReservedSlug(slug) {
		return "", ErrLinkNotFound
	}

	cachedURL, err := utils.GetSlugCache(ctx, s.rdb, slug)
	if err == nil && cachedURL != "" {
		go func(targetSlug string) {
			_ = s.linkRepo.IncrementClickCount(context.Background(), targetSlug)
		}(slug)
		return cachedURL, nil
	}

	link, err := s.linkRepo.GetLinkBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	if link == nil {
		return "", ErrLinkNotFound
	}

	_ = utils.SetSlugCache(ctx, s.rdb, link.Slug, link.OriginalURL, 24*time.Hour)

	go func(targetSlug string) {
		_ = s.linkRepo.IncrementClickCount(context.Background(), targetSlug)
	}(link.Slug)

	return link.OriginalURL, nil
}

func (s *linkService) CheckSlugAvailability(ctx context.Context, slug string) (bool, string, error) {
	slug = strings.TrimSpace(slug)
	if len(slug) < 3 || len(slug) > 50 {
		return false, "Slug must be between 3 and 50 characters", nil
	}
	if !utils.IsValidSlugFormat(slug) {
		return false, "Slug can only contain alphanumeric characters and hyphens", nil
	}
	if utils.IsReservedSlug(slug) {
		return false, "Slug cannot use reserved words (api, login, register, dashboard)", nil
	}
	exists, err := s.linkRepo.IsSlugExists(ctx, slug)
	if err != nil {
		return false, "", err
	}
	if exists {
		return false, "Slug already taken", nil
	}
	return true, "Slug is available", nil
}

func (s *linkService) generateUniqueSlug(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		randomSlug, err := utils.GenerateRandomString(6)
		if err != nil {
			return "", err
		}

		if utils.IsReservedSlug(randomSlug) {
			continue
		}

		exists, err := s.linkRepo.IsSlugExists(ctx, randomSlug)
		if err != nil {
			return "", err
		}
		if !exists {
			return randomSlug, nil
		}
	}
	return "", errors.New("failed to generate unique slug, please try again")
}
