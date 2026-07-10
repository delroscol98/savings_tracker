package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"

	"github.com/lib/pq"
)

func (a *ApiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := CreateUserParams{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	validatedParams, fieldsErrors := ValidateCreateUserParams(params)
	if fieldsErrors != nil {
		respondWithValidationError(w, http.StatusBadRequest, ValidationErrorBody{
			Error:  "Invalid parameters to create new user",
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

func (a *ApiConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	params := LoginParams{}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	validatedParams, fieldsErrors := ValidateLoginParams(params)
	if fieldsErrors != nil {
		respondWithValidationError(w, http.StatusBadRequest, ValidationErrorBody{
			Error:  "Incorrect email or password",
			Fields: fieldsErrors,
		})
		return
	}

	user, err := a.DatabaseQueries.Login(r.Context(), validatedParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusBadRequest, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unexpected database failure")
		return
	}

	match, _ := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if !match {
		respondWithError(w, http.StatusForbidden, "Incorrect email or password")
		return
	}

	respondWithJSON(
		w, http.StatusOK, User{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	)
}
