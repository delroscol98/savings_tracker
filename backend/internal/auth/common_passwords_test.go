package auth_test

import (
	"testing"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
)

func TestIsCommonPassword(t *testing.T) {
	tests := []struct {
		name             string
		password         string
		isCommonPassword bool
	}{
		{
			name:             "Known common password",
			password:         "password",
			isCommonPassword: true,
		},
		{
			name:             "Known common password with case insensitivity",
			password:         "PASSWORD",
			isCommonPassword: true,
		},
		{
			name:             "Complex password",
			password:         "ThisIsAComplexPasswordBecauseItIsLong",
			isCommonPassword: false,
		},
		{
			name:             "Empty password",
			password:         "",
			isCommonPassword: true,
		},
		{
			name:             "Common password with whitespace",
			password:         "       password         ",
			isCommonPassword: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCommonPassword := auth.IsCommonPassword(tt.password)
			if isCommonPassword != tt.isCommonPassword {
				t.Errorf(`
Expected password commonality: %v
Actual password commonality:   %v
`, tt.isCommonPassword, isCommonPassword)
			}
		})
	}
}
