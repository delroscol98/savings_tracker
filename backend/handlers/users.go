package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"

	"github.com/lib/pq"
)

func (a *ApiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type CreateUserRequestParams struct {
		Email string `json:"email"`
	}
	params := CreateUserRequestParams{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email cannot be empty")
		return
	}

	_, err = mail.ParseAddress(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid email")
		return
	}

	user, err := a.DatabaseQueries.CreateUser(r.Context(), params.Email)
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

	respondWithJSON(w, http.StatusCreated, user)
}
