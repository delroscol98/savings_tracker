package handlers

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
)

const (
	MIN_PASSWORD_LENGTH = 8
	MAX_PASSWORD_LENGTH = 128
)

func ValidateEmail(email string, fieldsErrors FieldErrors) (string, FieldErrors) {
	// Email validation
	if email == "" {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Email cannot be empty")
	}

	parsedAddress, err := mail.ParseAddress(email)
	if err != nil {
		fieldsErrors["email"] = append(fieldsErrors["email"], "Invalid email")
	} else {
		email = parsedAddress.Address
	}

	return email, fieldsErrors
}

func ValidatePassword(password string, fieldsErrors FieldErrors) FieldErrors {
	// Password validation
	if password == "" {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password cannot be empty")
	}

	if utf8.RuneCountInString(password) < MIN_PASSWORD_LENGTH {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password must be at least 8 characters")
	}

	if utf8.RuneCountInString(password) > MAX_PASSWORD_LENGTH {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password must be less than 128 characters in length")
	}

	if auth.IsCommonPassword(password) {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password is too common")
	}

	return fieldsErrors
}

func ValidateFullName(fullname string, fieldsErrors FieldErrors) FieldErrors {
	// Full name validation
	if fullname == "" {
		fieldsErrors["full_name"] = append(fieldsErrors["full_name"], "Full name cannot be empty")
	}

	return fieldsErrors
}

func ValidateCreateUserParams(params CreateUserParams) (CreateUserParams, FieldErrors) {
	originalParams := params
	fieldsErrors := make(FieldErrors)

	// Email validation
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	params.Email, fieldsErrors = ValidateEmail(params.Email, fieldsErrors)

	// Password validation
	params.Password = strings.ToLower(strings.TrimSpace(params.Password))
	fieldsErrors = ValidatePassword(params.Password, fieldsErrors)

	// Full name validation
	params.FullName = strings.ToLower(strings.TrimSpace(params.FullName))
	fieldsErrors = ValidateFullName(params.FullName, fieldsErrors)

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
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	params.Email, fieldsErrors = ValidateEmail(params.Email, fieldsErrors)

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

func ValidateResetResetPasswordParams(params ResetPasswordParams) (ResetPasswordParams, FieldErrors) {
	originalParams := params
	fieldsErrors := make(FieldErrors)

	// Password validation
	params.Password = strings.ToLower(strings.TrimSpace(params.Password))
	fieldsErrors = ValidatePassword(params.Password, fieldsErrors)

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return params, nil
	} else {
		return originalParams, fieldsErrors
	}
}
