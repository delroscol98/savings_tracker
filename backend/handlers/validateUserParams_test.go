package handlers_test

import (
	"testing"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/google/go-cmp/cmp"
)

func TestValidateUserParams(t *testing.T) {
	tests := []struct {
		name        string
		params      handlers.UserParams
		fieldErrors handlers.FieldErrors
	}{
		{
			name: "valid params",
			params: handlers.UserParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: nil,
		},
		{
			name: "valid params with name and address",
			params: handlers.UserParams{
				Email:    "Test <test@example.com>",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: nil,
		},
		{
			name: "empty email",
			params: handlers.UserParams{
				Email:    "",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: handlers.FieldErrors{
				"email": []string{"Email cannot be empty", "Invalid email"},
			},
		},
		{
			name: "invalid email and common password",
			params: handlers.UserParams{
				Email:    "invalidemail",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: handlers.FieldErrors{
				"email": []string{"Invalid email"},
			},
		},
		{
			name: "empty password",
			params: handlers.UserParams{
				Email:    "test@example.com",
				Password: "",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password cannot be empty", "Password must be at least 8 characters", "Password is too common"},
			},
		},
		{
			name: "too short password and common password",
			params: handlers.UserParams{
				Email:    "test@example.com",
				Password: "test",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password must be at least 8 characters", "Password is too common"},
			},
		},
		{
			name: "too long password",
			params: handlers.UserParams{
				Email:    "test@example.com",
				Password: "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890",
			},
			fieldErrors: handlers.FieldErrors{
				"password": []string{"Password must be less than 128 characters in length"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, fieldErrors := handlers.ValidateUserParams(tt.params)
			if !cmp.Equal(fieldErrors, tt.fieldErrors) {
				t.Errorf("FieldErrors structs do not match:\n%v", cmp.Diff(fieldErrors, tt.fieldErrors))
			}
		})
	}
}
