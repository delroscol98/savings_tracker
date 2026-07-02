package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"

	_ "embed"
)

type ErrorBody struct {
	Error string `json:"error"`
}

type ValidationErrorBody struct {
	Error  string      `json:"error"`
	Fields FieldErrors `json:"fields"`
}

type FieldErrors map[string][]string

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling JSON: %v", err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if code != http.StatusNoContent && code != http.StatusNotModified {
		num, err := w.Write(data)
		if err != nil {
			log.Printf("Error writing body: %v. Wrote %d bytes out of %d", err, num, len(data))
			return
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	log.Println(msg)

	params := ErrorBody{
		Error: msg,
	}

	respondWithJSON(w, code, params)
}

func respondWithValidationError(w http.ResponseWriter, code int, params ValidationErrorBody) {
	log.Println(params.Error)

	respondWithJSON(w, code, params)
}

func validateCreateUserParams(params CreateUserParams) (CreateUserParams, FieldErrors) {
	const MIN_PASSWORD_LENGTH = 8
	const MAX_PASSWORD_LENGTH = 128

	fieldsErrors := make(FieldErrors)

	// Email validation
	if params.Email == "" {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Email cannot be empty")
	}

	parsedAddress, err := mail.ParseAddress(params.Email)
	if err != nil {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Invalid email")
	} else {
		params.Email = parsedAddress.Address
	}

	params.Email = strings.ToLower(strings.TrimSpace(params.Email))

	// Password validation
	if params.Password == "" {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password cannot be empty")
	}

	if len(params.Password) < MIN_PASSWORD_LENGTH {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password must be at least 8 characters")
	}

	if len(params.Password) > MAX_PASSWORD_LENGTH {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password must be less than 128 characters in length")
	}

	if isCommonPassword(params.Password) {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password is too common")
	}

	// Check for any error messages
	if len(fieldsErrors["email"]) == 0 && len(fieldsErrors["password"]) == 0 {
		return params, nil
	} else {
		return CreateUserParams{}, fieldsErrors
	}
}
