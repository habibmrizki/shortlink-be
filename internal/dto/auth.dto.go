package dto

import "time"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"habib@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"ShortLink123!"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"habib@example.com"`
	Password string `json:"password" binding:"required" example:"ShortLink123!"`
}

type UserResponse struct {
	ID        int        `json:"id" example:"1"`
	Email     string     `json:"email" example:"habib@example.com"`
	Fullname  *string    `json:"fullname,omitempty" example:"Habib Muhammad Rizki"`
	AvatarURL *string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.png"`
	CreatedAt *time.Time `json:"created_at,omitempty" example:"2026-06-17T15:04:05Z"`
}

type LoginResult struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.signature"`
	User  UserResponse `json:"user"`
}
