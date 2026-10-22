package dto

import "time"

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	Slug        string `json:"slug,omitempty"`
	CustomSlug  string `json:"custom_slug,omitempty"`
}

type LinkResponse struct {
	ID          int        `json:"id"`
	OriginalURL string     `json:"original_url"`
	Slug        string     `json:"slug"`
	ShortURL    string     `json:"short_url"`
	ClickCount  int        `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type PaginatedLinkResponse struct {
	Data       []LinkResponse `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}
