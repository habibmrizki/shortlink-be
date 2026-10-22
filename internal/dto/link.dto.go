package dto

import "time"

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required" example:"https://www.linkedin.com/in/habib-rizki"`
	Slug        string `json:"slug,omitempty" example:"habib-profile"`
	CustomSlug  string `json:"custom_slug,omitempty" example:"habib-profile"`
}

type LinkResponse struct {
	ID          int        `json:"id" example:"1"`
	OriginalURL string     `json:"original_url" example:"https://www.linkedin.com/in/habib-rizki"`
	Slug        string     `json:"slug" example:"habib-profile"`
	ShortURL    string     `json:"short_url" example:"http://localhost:2901/habib-profile"`
	ClickCount  int        `json:"click_count" example:"42"`
	CreatedAt   time.Time  `json:"created_at" example:"2026-06-17T15:04:05Z"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	TotalItems int `json:"total_items" example:"45"`
	TotalPages int `json:"total_pages" example:"5"`
}

type PaginatedLinkResponse struct {
	Data       []LinkResponse `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}
