package models

import (
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Message struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Content   string         `json:"content"`
	MediaURLs pq.StringArray `json:"media_urls"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Reply struct {
	ID        string         `json:"id"`
	MessageID string         `json:"message_id"`
	UserID    string         `json:"user_id"`
	Content   string         `json:"content"`
	MediaURLs pq.StringArray `json:"media_urls"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type CreateMessageRequest struct {
	Content   string   `json:"content" binding:"required"`
	MediaURLs []string `json:"media_urls"`
}

type CreateReplyRequest struct {
	Content   string   `json:"content" binding:"required"`
	MediaURLs []string `json:"media_urls"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
