package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Queries interface {
	CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error)
	Login(ctx context.Context, email string) (database.User, error)
	CreatePasswordResetToken(ctx context.Context, params database.CreatePasswordResetTokenParams) (database.PasswordResetToken, error)
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (database.PasswordResetToken, error)
	ConsumePasswordResetToken(ctx context.Context, id uuid.UUID) error
	DeactivateUserTokens(ctx context.Context, userID uuid.UUID) error
	UpdateUserPassword(ctx context.Context, params database.UpdateUserPasswordParams) error
}

type UsersConfig struct {
	Queries                  Queries
	Database                 *sql.DB
	PasswordResetRateLimiter *ratelimit.RateLimiter
	LoginRateLimiter         *ratelimit.RateLimiter
	LoginThrottler           *ratelimit.LoginThrottler
	EmailSender              EmailSender
	JWTSecret                string
	BaseURL                  string
}

func (a *UsersConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := CreateUserParams{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	validatedParams, fieldsErrors := ValidateCreateUserParams(params)
	if fieldsErrors != nil {
		log.Print(fieldsErrors)
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to create new user",
			Fields: fieldsErrors,
		})
		return
	}

	hashedPW, err := auth.HashPassword(validatedParams.Password)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := a.Queries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          validatedParams.Email,
		HashedPassword: hashedPW,
		FullName:       validatedParams.FullName,
	})
	if err != nil {
		// PostgreSQL's unique violation code is 23505
		var pqe *pq.Error
		if errors.As(err, &pqe) && pqe.Code == "23505" {
			log.Print(err)
			response.RespondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error creating user")
		return
	}

	response.RespondWithJSON(
		w, http.StatusCreated, User{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	)
}

func (a *UsersConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	params := LoginParams{}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	validatedParams, fieldsErrors := ValidateLoginParams(params)
	if fieldsErrors != nil {
		log.Print(fieldsErrors)
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Incorrect email or password",
			Fields: fieldsErrors,
		})
		return
	}

	if !a.LoginRateLimiter.Allow(ratelimit.ClientIP(r)) {
		response.RespondWithError(w, http.StatusTooManyRequests, "Too many login attempts")
		return
	}

	if a.LoginThrottler.IsLockedOut(validatedParams.Email) {
		response.RespondWithError(w, http.StatusTooManyRequests, "Too many failed login attempts")
		return
	}

	user, err := a.Queries.Login(r.Context(), validatedParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Print(err)
			a.LoginThrottler.RecordFailure(validatedParams.Email)
			response.RespondWithError(w, http.StatusForbidden, "Incorrect email or password")
			return
		}
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "Unexpected database failure")
		return
	}

	match, _ := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if !match {
		a.LoginThrottler.RecordFailure(validatedParams.Email)
		response.RespondWithError(w, http.StatusForbidden, "Incorrect email or password")
		return
	}

	a.LoginThrottler.Clear(validatedParams.Email)

	token, err := auth.MakeJWT(user.ID, a.JWTSecret, time.Duration(validatedParams.ExpiresIn)*time.Second)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "Error creating JWT token")
		return
	}

	response.RespondWithJSON(
		w, http.StatusOK, User{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     token,
		},
	)
}
