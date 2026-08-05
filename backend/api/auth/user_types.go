package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
	Token          string    `json:"token"`
}

type CreateUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginParams struct {
	Email     string        `json:"email"`
	Password  string        `json:"password"`
	ExpiresIn int64 `json:"expires_in"` // seconds
}

type requestPasswordResetBody struct {
	Email string `json:"email"`
}

type ResetPasswordParams struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
