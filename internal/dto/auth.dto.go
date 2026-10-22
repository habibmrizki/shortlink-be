package dto

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	Fullname  *string `json:"fullname,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type LoginResult struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
