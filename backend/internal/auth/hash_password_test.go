package auth_test

import (
	"testing"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
)

func TestHashPassword(t *testing.T) {
	// First we create some hashed passwords
	password1 := "password1234!@#$"
	password2 := "password0987)(*&"
	password1Hash, err := auth.HashPassword(password1)
	if err != nil {
		t.Errorf(`
expected error: nil
actual error:   %v`, err)
	}
	password2Hash, err := auth.HashPassword(password2)
	if err != nil {
		t.Errorf(`
expected error: nil
actual error:   %v`, err)
	}

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          password1Hash,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "WrongPassword",
			hash:          password1Hash,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          password2Hash,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          password1Hash,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "InvalidHash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := auth.CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}
