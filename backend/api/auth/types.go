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
	ExpiresIn time.Duration `json:"expires_in"`
}

type requestPasswordResetbody struct {
	Email string `json:"email"`
}

type ResetPasswordParams struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type CreateGoalParams struct {
	Target   int32     `json:"target"`
	Deadline time.Time `json:"deadline"`
	UserId   uuid.UUID `json:"user_id"`
}

type Goal struct {
	Id        uuid.UUID `json:"id"`
	Target    int32     `json:"target"`
	Deadline  time.Time `json:"deadline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    uuid.UUID `json:"user_id"`
}
