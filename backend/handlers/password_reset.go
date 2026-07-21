package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/delroscol98/savings_tracker/backend/handlers/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

func (a *ApiConfig) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	body := requestPasswordResetbody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	// Email validation
	fieldsErrors := make(response.FieldErrors)
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	body.Email, fieldsErrors = ValidateEmail(body.Email, fieldsErrors)
	if len(fieldsErrors) != 0 {
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to reset password",
			Fields: fieldsErrors,
		})
		return
	}

	ip := "unknown"
	ipPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err == nil {
		ip = ipPort.Addr().String()
	}

	allowed := a.RateLimiter.Allow(ip)
	if !allowed {
		response.RespondWithError(w, http.StatusTooManyRequests, "Exceeded password reset limit")
		return
	}

	user, err := a.DatabaseQueries.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		log.Print("User not found")
		response.RespondWithJSON(w, http.StatusOK, struct {
			Message string `json:"message"`
		}{
			Message: "If the email exists, a reset link has been sent",
		})
		return
	}

	// Begin database transaction
	tx, err := a.Db.BeginTx(r.Context(), nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating db transaction: %s", err)
		response.RespondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	err = qtx.DeactivateUserTokens(r.Context(), user.ID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Error deactivating user tokens")
		return
	}

	token, err := auth.GenerateResetToken()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tokenHash := auth.HashToken(token)

	_, err = qtx.CreatePasswordResetToken(r.Context(), database.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
	})
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Error storing password reset token")
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		errMsg := fmt.Sprintf("Error committing transaction: %s", err)
		response.RespondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	link := fmt.Sprintf("http://%v/reset-password?token=%v", r.Host, token)
	log.Printf("Password reset link: %v", link)
	err = a.EmailSender.Send(user.Email, "Savings-Tracker: Reset Your Password", fmt.Sprintf(`
	<p>Click the following <a href="%v">link</a> to reset your password.</p>
	`, link))
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Error sending reset password link")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "If the email exists, a reset link has been sent",
	})
}

func (a *ApiConfig) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	body := ResetPasswordParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	body, fieldsErrors := ValidateResetResetPasswordParams(body)
	if fieldsErrors != nil {
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to reset password",
			Fields: fieldsErrors,
		})
		return
	}

	hashToken := auth.HashToken(body.Token)

	passwordResetToken, err := a.DatabaseQueries.GetPasswordResetTokenByHash(r.Context(), hashToken)
	if err != nil || time.Now().After(passwordResetToken.ExpiresAt) {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	if passwordResetToken.ConsumedAt.Valid && time.Now().After(passwordResetToken.ConsumedAt.Time) {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	pwHash, err := auth.HashPassword(body.Password)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Begin database transaction
	tx, err := a.Db.BeginTx(r.Context(), nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating db transaction: %s", err)
		response.RespondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	err = qtx.UpdateUserPassword(r.Context(), database.UpdateUserPasswordParams{
		ID:             passwordResetToken.UserID,
		HashedPassword: pwHash,
	})
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Error updating password")
		return
	}

	err = qtx.ConsumePasswordResetToken(r.Context(), passwordResetToken.ID)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Error consuming password reset token")
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		errMsg := fmt.Sprintf("Error committing transaction: %s", err)
		response.RespondWithError(w, http.StatusInternalServerError, errMsg)
		return
	}

	response.RespondWithJSON(
		w, http.StatusOK, struct {
			Message string `json:"message"`
		}{
			Message: "Password successfully reset",
		},
	)
}
