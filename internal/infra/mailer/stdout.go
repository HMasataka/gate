package mailer

import (
	"context"
	"log/slog"
)

type StdoutMailer struct{}

func NewStdoutMailer() *StdoutMailer {
	return &StdoutMailer{}
}

func (m *StdoutMailer) SendEmailVerification(ctx context.Context, email, token string) error {
	slog.Info("email verification",
		"to", email,
		"token", token,
		"subject", "Verify your email address",
	)
	return nil
}

func (m *StdoutMailer) SendPasswordReset(ctx context.Context, email, token string) error {
	slog.Info("password reset",
		"to", email,
		"token", token,
		"subject", "Reset your password",
	)
	return nil
}
