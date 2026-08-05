package api_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func extractToken(t *testing.T, logOutput string) string {
	t.Helper()
	re := regexp.MustCompile(`token=([a-f0-9]+)`)
	matches := re.FindStringSubmatch(logOutput)
	if len(matches) < 2 {
		t.Fatalf("Token not found in log output: %v", logOutput)
	}
	return matches[1]
}

func TestFullPasswordResetFlow_Integration(t *testing.T) {
	testRateLimiter.Reset()

	// Create user
	createUserBody := strings.NewReader(`
{
	"email": "johnsmith@testexample.com",
	"password": "AnotherTestPassword",
	"full_name": "John Smith"
}
	`)
	resp, err := http.Post(testServer.URL+"/api/users", "application/json", createUserBody)
	if err != nil {
		t.Fatalf("Error creating user in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf(`
Creating a new user:

Expected status code: %v
Actual status code:   %v
`, http.StatusCreated, resp.StatusCode)
	}

	// Forgot password

	forgotPasswordBody := strings.NewReader(`
{
	"email": "johnsmith@testexample.com"
}
	`)
	resp, err = http.Post(testServer.URL+"/api/forgot-password", "application/json", forgotPasswordBody)
	if err != nil {
		t.Fatalf("Error creating reset password link in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`
Creating a reset password link:

Expected status code: %v
Actual status code:   %v
`, http.StatusOK, resp.StatusCode)
	}

	if len(mockSender.Sent) != 1 {
		t.Fatalf(`
Expected number of sent emails: 1
Actual number of sent emails:   %v
`, len(mockSender.Sent))
	}

	if mockSender.Sent[0].To != "johnsmith@testexample.com" {
		t.Fatalf(`
Expected To email: johnsmith@testexample.com
Actual to email:   %v
`, mockSender.Sent[0].To)
	}

	token := extractToken(t, mockSender.Sent[0].Html)

	// Reset Password
	resetPasswordBody := strings.NewReader(fmt.Sprintf(`
{
	"token": "%v",
	"password": "DifferentTestPassword"
}
	`, token))

	resp, err = http.Post(testServer.URL+"/api/reset-password", "application/json", resetPasswordBody)
	if err != nil {
		t.Fatalf("Error resetting password in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`
Resetting password:

Expected status code: %v
Actual status code:   %v
`, http.StatusOK, resp.StatusCode)
	}

	// Login with new password "DifferentTestPassword"
	loginBody := strings.NewReader(`
{
	"email": "johnsmith@testexample.com",
	"password": "DifferentTestPassword"
}
`)

	resp, err = http.Post(testServer.URL+"/api/login", "application/json", loginBody)
	if err != nil {
		t.Fatalf("Error logging in with new password in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`
Logging in:

Expected status code: %v,
Actual status code:   %v,
`, http.StatusOK, resp.StatusCode)
	}

	// Login with old password "AnotherTestPassword"
	loginBody = strings.NewReader(`
{
	"email": "johnsmith@testexample.com",
	"password": "AnotherTestPassword"
}
`)

	resp, err = http.Post(testServer.URL+"/api/login", "application/json", loginBody)
	if err != nil {
		t.Fatalf("Error logging in with old password in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf(`
Logging in:

Expected status code: %v
Actual status code:   %v
`, http.StatusForbidden, resp.StatusCode)
	}
}
