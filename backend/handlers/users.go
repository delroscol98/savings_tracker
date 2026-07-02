package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"

	"github.com/lib/pq"
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
}

func (a *ApiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := CreateUserParams{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	validatedParams, fieldsErrors := validateCreateUserParams(params)
	if fieldsErrors != nil {
		respondWithValidationError(w, http.StatusBadRequest, ValidationErrorBody{
			Error:  "Invalid parameters for creating a user",
			Fields: fieldsErrors,
		})
		return
	}

	hashedPW, err := auth.HashPassword(validatedParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := a.DatabaseQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          validatedParams.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		// PostgreSQL's unique violation code is 23505
		var pqe *pq.Error
		if errors.As(err, &pqe) && pqe.Code == "23505" {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusBadRequest, "Error creating user")
		return
	}

	respondWithJSON(
		w, http.StatusCreated, User{
			Id:             user.ID,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
			Email:          user.Email,
			HashedPassword: user.HashedPassword,
		},
	)
}
