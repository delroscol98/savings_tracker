package handlers

import (
	"net/mail"
	"strings"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
)

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

	if auth.IsCommonPassword(params.Password) {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password is too common")
	}

	// Check for any error messages
	if len(fieldsErrors["email"]) == 0 && len(fieldsErrors["password"]) == 0 {
		return params, nil
	} else {
		return CreateUserParams{}, fieldsErrors
	}
}
