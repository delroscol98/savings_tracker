package auth

import (
	"net/mail"
	"unicode/utf8"

	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

const (
	MIN_PASSWORD_LENGTH = 8
	MAX_PASSWORD_LENGTH = 128
)

func ValidateEmail(email string, fieldsErrors response.FieldErrors) (string, response.FieldErrors) {
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

func ValidatePassword(password string, fieldsErrors response.FieldErrors) response.FieldErrors {
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

	if IsCommonPassword(password) {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password is too common")
	}

	return fieldsErrors
}

func ValidateFullName(fullname string, fieldsErrors response.FieldErrors) response.FieldErrors {
	// Full name validation
	if fullname == "" {
		fieldsErrors["full_name"] = append(fieldsErrors["full_name"], "Full name cannot be empty")
	}

	return fieldsErrors
}
