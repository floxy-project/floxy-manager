package contract

import "context"

// Emailer defines the interface for sending emails.
type Emailer interface {
	// SendResetPasswordEmail sends a password reset email with a token.
	SendResetPasswordEmail(ctx context.Context, email, token string) error
	// Send2FACodeEmail sends a 2FA code email for the specified action (disable/reset).
	Send2FACodeEmail(ctx context.Context, email, code, action string) error
}
