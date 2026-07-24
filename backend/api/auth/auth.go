package auth

import (
	"context"
	"database/sql"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
	"github.com/google/uuid"
	"github.com/resend/resend-go/v3"
)

type Queries interface {
	CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error)
	Login(ctx context.Context, email string) (database.User, error)
	CreatePasswordResetToken(ctx context.Context, params database.CreatePasswordResetTokenParams) (database.PasswordResetToken, error)
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (database.PasswordResetToken, error)
	ConsumePasswordResetToken(ctx context.Context, id uuid.UUID) error
	DeactivateUserTokens(ctx context.Context, userID uuid.UUID) error
	UpdateUserPassword(ctx context.Context, params database.UpdateUserPasswordParams) error
}

type EmailSender interface {
	Send(to, subject, html string) error
}

type AuthConfig struct {
	Queries     Queries
	Database    *sql.DB
	RateLimiter *ratelimit.RateLimiter
	EmailSender EmailSender
}

type ResendSender struct {
	Client *resend.Client
	From   string
}

func (rs *ResendSender) Send(to, subject, html string) error {
	params := &resend.SendEmailRequest{
		From:    rs.From,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	_, err := rs.Client.Emails.Send(params)
	return err
}
