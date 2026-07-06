package handlers_test

import (
	"testing"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/google/go-cmp/cmp"
)

func TestValidateCreateUserParams(t *testing.T) {
	tests := []struct {
		name        string
		params      handlers.CreateUserParams
		fieldErrors handlers.FieldErrors
	}{
		{
			name: "valid params",
			params: handlers.CreateUserParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: nil,
		},
		{
			name: "valid params with name and address",
			params: handlers.CreateUserParams{
				Email:    "Test <test@example.com>",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: nil,
		},
		{
			name: "empty email",
			params: handlers.CreateUserParams{
				Email:    "",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: handlers.FieldErrors{
				"email": []string{"Email cannot be empty", "Invalid email"},
			},
		},
		{
			name: "invalid email and common password",
			params: handlers.CreateUserParams{
				Email:    "invalidemail",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: handlers.FieldErrors{
				"email": []string{"Invalid email"},
			},
		},
		{
			name: "empty password",
			params: handlers.CreateUserParams{
				Email:    "test@example.com",
				Password: "",
				FullName: "John Smith",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password cannot be empty", "Password must be at least 8 characters", "Password is too common"},
			},
		},
		{
			name: "too short password and common password",
			params: handlers.CreateUserParams{
				Email:    "test@example.com",
				Password: "test",
				FullName: "John Smith",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password must be at least 8 characters", "Password is too common"},
			},
		},
		{
			name: "too long password",
			params: handlers.CreateUserParams{
				Email:    "test@example.com",
				Password: "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890",
				FullName: "John Smith",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password must be less than 128 characters in length"},
			},
		},
		{
			name: "Empty name",
			params: handlers.CreateUserParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
				FullName: "",
			},
			fieldErrors: handlers.FieldErrors{
				"full_name": []string{"Full name cannot be empty"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, fieldErrors := handlers.ValidateCreateUserParams(tt.params)
			if !cmp.Equal(fieldErrors, tt.fieldErrors) {
				t.Errorf("FieldErrors structs do not match:\n%v", cmp.Diff(fieldErrors, tt.fieldErrors))
			}
		})
	}
}

func TestValidateLoginParams(t *testing.T) {
	tests := []struct {
		name        string
		params      handlers.LoginParams
		fieldErrors handlers.FieldErrors
	}{
		{
			name: "valid params",
			params: handlers.LoginParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: nil,
		},
		{
			name: "empty email",
			params: handlers.LoginParams{
				Email:    "",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: handlers.FieldErrors{
				"email": []string{"Email cannot be empty", "Invalid email"},
			},
		},
		{
			name: "invalid email",
			params: handlers.LoginParams{
				Email:    "invalidemail",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: handlers.FieldErrors{
				"email": []string{"Invalid email"},
			},
		},
		{
			name: "empty password",
			params: handlers.LoginParams{
				Email:    "test@example.com",
				Password: "",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password cannot be empty"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, fieldErrors := handlers.ValidateLoginParams(tt.params)
			if !cmp.Equal(fieldErrors, tt.fieldErrors) {
				t.Errorf("FieldErrors structs do not match:\n%v", cmp.Diff(fieldErrors, tt.fieldErrors))
			}
		})
	}
}
