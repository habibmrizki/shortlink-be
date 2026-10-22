package models

import "time"

type Link struct {
	Id          int        `json:"id" db:"id"`
	UserId      int        `json:"user_id" db:"user_id"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	Slug        string     `json:"slug" db:"slug"`
	ClickCount  int        `json:"click_count" db:"click_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
