package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/response"

	"github.com/lib/pq"
)

func (a *AuthConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	hashedPW, err := HashPassword(validatedParams.Password)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := a.Queries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          validatedParams.Email,
		HashedPassword: hashedPW,
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
			Id:             user.ID,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
			Email:          user.Email,
			HashedPassword: user.HashedPassword,
		},
	)
}

func (a *AuthConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := a.Queries.Login(r.Context(), validatedParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Print(err)
			response.RespondWithError(w, http.StatusBadRequest, "User not found")
			return
		}
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "Unexpected database failure")
		return
	}

	match, _ := CheckPasswordHash(params.Password, user.HashedPassword)
	if !match {
		response.RespondWithError(w, http.StatusForbidden, "Incorrect email or password")
		return
	}

	token, err := MakeJWT(user.ID, a.JWTSecret, params.ExpiresIn)
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
