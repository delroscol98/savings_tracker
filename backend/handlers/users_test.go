package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

func TestCreateUserHandler_Valid(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{
			users: make(map[string]database.User),
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"email": "test@example.com"}`))
	r.Header.Set("Content-Type", "application/json")
	api.CreateUserHandler(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf(`
Expected status code: 201
Actual status code:   %v
`, w.Code)
	}

	user := database.User{}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&user)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v
`, err)
	}

	if user.Email != "test@example.com" {
		t.Errorf(`
Expected email: test@example.com
Actual email:   %v
`, user.Email)
	}

	if user.ID == uuid.Nil {
		t.Error("user ID should NOT be UUID zero-value")
	}

	if user.CreatedAt.IsZero() {
		t.Error("user created_at should NOT be timestamp zero-value")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("user updated_at should NOT be timestamp zero-value")
	}
}

func TestCreateUserHandler_EmptyEmail(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{
			users: make(map[string]database.User),
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"email": ""}`))
	r.Header.Set("Content-Type", "application/json")
	api.CreateUserHandler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf(`
Expected status code: 400
Actual status code:   %v
`, w.Code)
	}

	body := make(map[string]interface{})
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&body)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v
`, err)
	}

	if body["error"] != "Email cannot be empty" {
		t.Errorf(`
Expected error: Email cannot be empty
Actual error:   %v
`, body["error"])
	}
}

func TestCreateUserHandler_InvalidEmail(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{
			users: make(map[string]database.User),
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"email": "ThisIsAnInvalidEmail"}`))
	r.Header.Set("Content-Type", "application/json")
	api.CreateUserHandler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf(`
Expected status code: 400
Actual status code:   %v
`, w.Code)
	}

	body := make(map[string]interface{})
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&body)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v
`, err)
	}

	if body["error"] != "Invalid email" {
		t.Errorf(`
Expected error: Invalid email
Actual error:   %v
`, body["error"])
	}
}

func TestCreateUserHandler_DuplicateEmail(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{
			users: map[string]database.User{
				"test@example.com": {
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Email:     "test@example.com",
				},
			},
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"email": "test@example.com"}`))
	r.Header.Set("Content-Type", "application/json")
	api.CreateUserHandler(w, r)

	if w.Code != http.StatusConflict {
		t.Errorf(`
Expected status code: 409
Actual status code:   %v
`, w.Code)
	}

	body := make(map[string]interface{})
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&body)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v
`, err)
	}

	if body["error"] != "Email already exists" {
		t.Errorf(`
Expected error: Email already exists
Actual error:   %v
`, body["error"])
	}
}

func TestCreateUserHandler_EmptyBody(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{
			users: make(map[string]database.User),
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	r.Header.Set("Content-Type", "application/json")
	api.CreateUserHandler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf(`
Expected status code: 400
Actual status code:   %v
`, w.Code)
	}

	body := make(map[string]interface{})
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&body)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v
`, err)
	}

	if body["error"] != "Error decoding body" {
		t.Errorf(`
Expected error: Error decoding body
Actual error:   %v
`, body["error"])
	}
}

func TestCreateUserHandler_DBError(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{
			CreateUserErr: errors.New("connection refused"),
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"email": "test@example.com"}`))
	r.Header.Set("Content-Type", "application/json")
	api.CreateUserHandler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf(`
Expected status code: 400
Actual status code:   %v
`, w.Code)
	}

	body := make(map[string]interface{})
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&body)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v
`, err)
	}

	if body["error"] != "Error creating user" {
		t.Errorf(`
Expected error: Error creating user
Actual error:   %v
`, body["error"])
	}
}
