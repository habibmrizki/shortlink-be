package repositories

import (
	"context"

	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLinkNotFound = utils.ErrLinkNotFound
)

type LinkRepository interface {
	CreateLink(ctx context.Context, link *models.Link) (*models.Link, error)
	IsSlugExists(ctx context.Context, slug string) (bool, error)
	GetLinkBySlug(ctx context.Context, slug string) (*models.Link, error)
	GetLinkByIDAndUserID(ctx context.Context, id, userID int) (*models.Link, error)
	GetLinksByUserID(ctx context.Context, userID int) ([]models.Link, error)
	GetLinksByUserIDWithPagination(ctx context.Context, userID int, page, limit int, showDeleted bool) ([]models.Link, int, error)
	DeleteLink(ctx context.Context, linkID, userID int) error
	IncrementClickCount(ctx context.Context, slug string) error
}

type linkRepository struct {
	db *pgxpool.Pool
}

func NewLinkRepository(db *pgxpool.Pool) LinkRepository {
	return &linkRepository{db: db}
}

// CreateLink creates a new link using a database transaction
func (r *linkRepository) CreateLink(ctx context.Context, link *models.Link) (*models.Link, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO links (user_id, original_url, slug)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, original_url, slug, click_count, created_at
	`
	var created models.Link
	err = tx.QueryRow(ctx, query, link.UserId, link.OriginalURL, link.Slug).Scan(
		&created.Id,
		&created.UserId,
		&created.OriginalURL,
		&created.Slug,
		&created.ClickCount,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &created, nil
}
