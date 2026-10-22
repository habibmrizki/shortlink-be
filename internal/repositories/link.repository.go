package repositories

import (
	"context"
	"errors"

	"github.com/habibmrizki/shorlink-be/internal/models"
	"github.com/habibmrizki/shorlink-be/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLinkNotFound = utils.ErrLinkNotFound
)

type LinkRepository interface {
	CreateLink(ctx context.Context, link *models.Link) (*models.Link, error)
	UpdateLink(ctx context.Context, link *models.Link) (*models.Link, error)
	IsSlugExists(ctx context.Context, slug string) (bool, error)
	IsSlugExistsExcludingID(ctx context.Context, slug string, excludeID int) (bool, error)
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

// UpdateLink updates a link using a database transaction
func (r *linkRepository) UpdateLink(ctx context.Context, link *models.Link) (*models.Link, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE links
		SET original_url = $1, slug = $2
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
		RETURNING id, user_id, original_url, slug, click_count, created_at, deleted_at
	`
	var updated models.Link
	err = tx.QueryRow(ctx, query, link.OriginalURL, link.Slug, link.Id, link.UserId).Scan(
		&updated.Id,
		&updated.UserId,
		&updated.OriginalURL,
		&updated.Slug,
		&updated.ClickCount,
		&updated.CreatedAt,
		&updated.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *linkRepository) IsSlugExists(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM links WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *linkRepository) IsSlugExistsExcludingID(ctx context.Context, slug string, excludeID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM links WHERE slug = $1 AND id != $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, slug, excludeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *linkRepository) GetLinkBySlug(ctx context.Context, slug string) (*models.Link, error) {
	query := `
		SELECT id, user_id, original_url, slug, click_count, created_at, deleted_at
		FROM links
		WHERE slug = $1 AND deleted_at IS NULL
	`
	var link models.Link
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&link.Id,
		&link.UserId,
		&link.OriginalURL,
		&link.Slug,
		&link.ClickCount,
		&link.CreatedAt,
		&link.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) GetLinkByIDAndUserID(ctx context.Context, id, userID int) (*models.Link, error) {
	query := `
		SELECT id, user_id, original_url, slug, click_count, created_at, deleted_at
		FROM links
		WHERE id = $1 AND user_id = $2
	`
	var link models.Link
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&link.Id,
		&link.UserId,
		&link.OriginalURL,
		&link.Slug,
		&link.ClickCount,
		&link.CreatedAt,
		&link.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) GetLinksByUserID(ctx context.Context, userID int) ([]models.Link, error) {
	query := `
		SELECT id, user_id, original_url, slug, click_count, created_at, deleted_at
		FROM links
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]models.Link, 0)
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.Id,
			&link.UserId,
			&link.OriginalURL,
			&link.Slug,
			&link.ClickCount,
			&link.CreatedAt,
			&link.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (r *linkRepository) GetLinksByUserIDWithPagination(ctx context.Context, userID int, page, limit int, showDeleted bool) ([]models.Link, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var countQuery string
	var dataQuery string

	if showDeleted {
		countQuery = `SELECT COUNT(*) FROM links WHERE user_id = $1 AND deleted_at IS NOT NULL`
		dataQuery = `
			SELECT id, user_id, original_url, slug, click_count, created_at, deleted_at
			FROM links
			WHERE user_id = $1 AND deleted_at IS NOT NULL
			ORDER BY deleted_at DESC
			LIMIT $2 OFFSET $3
		`
	} else {
		countQuery = `SELECT COUNT(*) FROM links WHERE user_id = $1 AND deleted_at IS NULL`
		dataQuery = `
			SELECT id, user_id, original_url, slug, click_count, created_at, deleted_at
			FROM links
			WHERE user_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
	}

	var totalItems int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, dataQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	links := make([]models.Link, 0)
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.Id,
			&link.UserId,
			&link.OriginalURL,
			&link.Slug,
			&link.ClickCount,
			&link.CreatedAt,
			&link.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		links = append(links, link)
	}
	return links, totalItems, nil
}

func (r *linkRepository) DeleteLink(ctx context.Context, linkID, userID int) error {
	query := `
		UPDATE links
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, linkID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrLinkNotFound
	}
	return nil
}

func (r *linkRepository) IncrementClickCount(ctx context.Context, slug string) error {
	query := `
		UPDATE links
		SET click_count = click_count + 1
		WHERE slug = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, slug)
	return err
}
