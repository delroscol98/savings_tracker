package users

import (
	"strings"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

const DEFAULT_LOGIN_EXPIRY_SECONDS = 3600

func ValidateCreateUserParams(params CreateUserParams) (CreateUserParams, response.FieldErrors) {
	originalParams := params
	fieldsErrors := make(response.FieldErrors)

	// Email validation
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	params.Email, fieldsErrors = auth.ValidateEmail(params.Email, fieldsErrors)

	// Password validation
	params.Password = strings.TrimSpace(params.Password)
	fieldsErrors = auth.ValidatePassword(params.Password, fieldsErrors)

	// Full name validation
	params.FullName = strings.ToLower(strings.TrimSpace(params.FullName))
	fieldsErrors = auth.ValidateFullName(params.FullName, fieldsErrors)

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return params, nil
	} else {
		return originalParams, fieldsErrors
	}
}

func ValidateLoginParams(params LoginParams) (LoginParams, response.FieldErrors) {
	originalParams := params
	fieldsErrors := make(response.FieldErrors)

	// Email validation
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	params.Email, fieldsErrors = auth.ValidateEmail(params.Email, fieldsErrors)

	// Password validation
	params.Password = strings.TrimSpace(params.Password)
	if params.Password == "" {
		fieldsErrors["password"] = append(fieldsErrors["password"], "Password cannot be empty")
	}

	// JWT Duration
	if params.ExpiresIn < 0 {
		fieldsErrors["expires_in"] = append(fieldsErrors["expires_in"], "Expires in cannot be negative")
	}
	if params.ExpiresIn == 0 {
		params.ExpiresIn = DEFAULT_LOGIN_EXPIRY_SECONDS
	}

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return params, nil
	} else {
		return originalParams, fieldsErrors
	}
}

func ValidateResetPasswordParams(params ResetPasswordParams) (ResetPasswordParams, response.FieldErrors) {
	originalParams := params
	fieldsErrors := make(response.FieldErrors)

	// Password validation
	params.Password = strings.TrimSpace(params.Password)
	fieldsErrors = auth.ValidatePassword(params.Password, fieldsErrors)

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return params, nil
	} else {
		return originalParams, fieldsErrors
	}
}
