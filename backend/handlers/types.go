package handlers

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
}

type CreateUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type requestPasswordResetbody struct {
	Email string `json:"email"`
}

type ResetPasswordParams struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
