package auth_test

import (
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
	"github.com/google/go-cmp/cmp"
)

func TestValidateCreateUserParams(t *testing.T) {
	tests := []struct {
		name        string
		params      auth.CreateUserParams
		fieldErrors response.FieldErrors
	}{
		{
			name: "valid params",
			params: auth.CreateUserParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: nil,
		},
		{
			name: "valid params with name and address",
			params: auth.CreateUserParams{
				Email:    "Test <test@example.com>",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: nil,
		},
		{
			name: "empty email",
			params: auth.CreateUserParams{
				Email:    "",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: response.FieldErrors{
				"email": []string{"Email cannot be empty", "Invalid email"},
			},
		},
		{
			name: "invalid email and common password",
			params: auth.CreateUserParams{
				Email:    "invalidemail",
				Password: "ThisIsATestPassword",
				FullName: "John Smith",
			},
			fieldErrors: response.FieldErrors{
				"email": []string{"Invalid email"},
			},
		},
		{
			name: "empty password",
			params: auth.CreateUserParams{
				Email:    "test@example.com",
				Password: "",
				FullName: "John Smith",
			},
			fieldErrors: response.FieldErrors{
				"password": []string{"Password cannot be empty", "Password must be at least 8 characters", "Password is too common"},
			},
		},
		{
			name: "too short password and common password",
			params: auth.CreateUserParams{
				Email:    "test@example.com",
				Password: "test",
				FullName: "John Smith",
			},
			fieldErrors: response.FieldErrors{
				"password": []string{"Password must be at least 8 characters", "Password is too common"},
			},
		},
		{
			name: "too long password",
			params: auth.CreateUserParams{
				Email:    "test@example.com",
				Password: "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890",
				FullName: "John Smith",
			},
			fieldErrors: response.FieldErrors{
				"password": []string{"Password must be less than 128 characters in length"},
			},
		},
		{
			name: "Empty name",
			params: auth.CreateUserParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
				FullName: "",
			},
			fieldErrors: response.FieldErrors{
				"full_name": []string{"Full name cannot be empty"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, fieldErrors := auth.ValidateCreateUserParams(tt.params)
			if !cmp.Equal(fieldErrors, tt.fieldErrors) {
				t.Errorf("FieldErrors structs do not match:\n%v", cmp.Diff(fieldErrors, tt.fieldErrors))
			}
		})
	}
}

func TestValidateLoginParams(t *testing.T) {
	tests := []struct {
		name        string
		params      auth.LoginParams
		wantParams  *auth.LoginParams
		fieldErrors response.FieldErrors
	}{
		{
			name: "valid params",
			params: auth.LoginParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: nil,
		},
		{
			name: "default expires_in applied",
			params: auth.LoginParams{
				Email:    "test@example.com",
				Password: "ThisIsATestPassword",
			},
			wantParams: &auth.LoginParams{
				Email:     "test@example.com",
				Password:  "ThisIsATestPassword",
				ExpiresIn: auth.DEFAULT_LOGIN_EXPIRY_SECONDS,
			},
			fieldErrors: nil,
		},
		{
			name: "negative expires_in",
			params: auth.LoginParams{
				Email:     "test@example.com",
				Password:  "ThisIsATestPassword",
				ExpiresIn: -1,
			},
			fieldErrors: response.FieldErrors{
				"expires_in": []string{"Expires in cannot be negative"},
			},
		},
		{
			name: "empty email",
			params: auth.LoginParams{
				Email:    "",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: response.FieldErrors{
				"email": []string{"Email cannot be empty", "Invalid email"},
			},
		},
		{
			name: "invalid email",
			params: auth.LoginParams{
				Email:    "invalidemail",
				Password: "ThisIsATestPassword",
			},
			fieldErrors: response.FieldErrors{
				"email": []string{"Invalid email"},
			},
		},
		{
			name: "empty password",
			params: auth.LoginParams{
				Email:    "test@example.com",
				Password: "",
			},
			fieldErrors: response.FieldErrors{
				"password": []string{"Password cannot be empty"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParams, fieldErrors := auth.ValidateLoginParams(tt.params)
			if !cmp.Equal(fieldErrors, tt.fieldErrors) {
				t.Errorf("FieldErrors structs do not match:\n%v", cmp.Diff(fieldErrors, tt.fieldErrors))
			}
			if tt.wantParams != nil && !cmp.Equal(gotParams, *tt.wantParams) {
				t.Errorf("Params structs do not match:\n%v", cmp.Diff(gotParams, *tt.wantParams))
			}
		})
	}
}

func TestValidateGoalParams(t *testing.T) {
	tests := []struct {
		name        string
		params      auth.GoalFields
		fieldErrors response.FieldErrors
	}{
		{
			name: "Valid Params",
			params: auth.GoalFields{
				Target:   1000,
				Deadline: time.Now().Add(time.Hour),
			},
			fieldErrors: nil,
		},
		{
			name: "Invalid target",
			params: auth.GoalFields{
				Target:   -1000,
				Deadline: time.Now().Add(time.Hour),
			},
			fieldErrors: response.FieldErrors{
				"target": []string{"Goal target cannot be negative"},
			},
		},
		{
			name: "Invalid deadline",
			params: auth.GoalFields{
				Target:   1000,
				Deadline: time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC),
			},
			fieldErrors: response.FieldErrors{
				"deadline": []string{"Deadline cannot be in the past"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldErrors := auth.ValidateGoalParams(tt.params)
			if !cmp.Equal(fieldErrors, tt.fieldErrors) {
				t.Errorf("FieldErrors structs do not match:\n%v", cmp.Diff(fieldErrors, tt.fieldErrors))
			}
		})
	}
}
