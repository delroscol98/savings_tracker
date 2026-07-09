package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/netip"
	"strings"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

func (a *ApiConfig) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	type requestPasswordResetbody struct {
		Email string `json:"email"`
	}

	body := requestPasswordResetbody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	// Email validation
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email cannot be empty")
		return
	}

	parsedAddress, err := mail.ParseAddress(body.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid email")
		return
	}
	body.Email = parsedAddress.Address

	ip := "unknown"
	ipPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err == nil {
		ip = ipPort.Addr().String()
	}

	allowed := a.RateLimiter.Allow(ip)
	if !allowed {
		respondWithError(w, http.StatusTooManyRequests, "Exceeded password reset limit")
		return
	}

	user, err := a.DatabaseQueries.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "User not found")
		return
	}

	// Begin database transaction
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating db transaction: %s", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	err = qtx.DeactivateUserTokens(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error deactivating user tokens")
		return
	}

	token, err := auth.GenerateResetToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tokenHash := auth.HashToken(token)

	_, err = qtx.CreatePasswordResetToken(r.Context(), database.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error storing password reset token")
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		errMsg := fmt.Sprintf("Error committing transaction: %s", err)
		respondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	link := fmt.Sprintf("http://%v/reset-password?token=%v", r.Host, token)
	log.Printf("Password reset link: %v", link)

	respondWithJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "If the email exists, a reset link has been sent",
	})
}
