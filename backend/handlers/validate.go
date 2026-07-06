package handlers

import (
	"net/mail"
	"strings"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
)

const (
	MIN_PASSWORD_LENGTH = 8
	MAX_PASSWORD_LENGTH = 128
)

func ValidateCreateUserParams(params CreateUserParams) (CreateUserParams, FieldErrors) {
	originalParams := params
	fieldsErrors := make(FieldErrors)

	// Email validation
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	if params.Email == "" {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Email cannot be empty")
	}

	parsedAddress, err := mail.ParseAddress(params.Email)
	if err != nil {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Invalid email")
	} else {
		params.Email = parsedAddress.Address
	}

	// Password validation
	params.Password = strings.ToLower(strings.TrimSpace(params.Password))
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

	// Full name validation
	params.FullName = strings.ToLower(strings.TrimSpace(params.FullName))
	if params.FullName == "" {
		fieldsErrors["full_name"] = append(fieldsErrors["full_name"], "Full name cannot be empty")
	}

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return params, nil
	} else {
		return originalParams, fieldsErrors
	}
}

func ValidateLoginParams(params LoginParams) (LoginParams, FieldErrors) {
	originalParams := params
	fieldsErrors := make(FieldErrors)

	// Email validation
	params.Email = strings.TrimSpace(params.Email)
	if params.Email == "" {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Email cannot be empty")
	}

	parsedAddress, err := mail.ParseAddress(params.Email)
	if err != nil {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Invalid email")
	} else {
		params.Email = parsedAddress.Address
	}

	// Password validation
	params.Password = strings.TrimSpace(params.Password)
	if params.Password == "" {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password cannot be empty")
	}

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return params, nil
	} else {
		return originalParams, fieldsErrors
	}
}
