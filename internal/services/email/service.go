package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"path"
	"strings"
	"time"

	"github.com/rom8726/floxy-manager/internal/contract"
)

// Config holds email service configuration.
type Config struct {
	SMTPHost      string
	Username      string
	Password      string
	CertFile      string
	KeyFile       string
	AllowInsecure bool
	UseTLS        bool
	BaseURL       string
	From          string
}

// Service implements contract.Emailer.
type Service struct {
	config Config
}

// New creates a new email service.
func New(config *Config) *Service {
	return &Service{
		config: *config,
	}
}

var _ contract.Emailer = (*Service)(nil)

// SendResetPasswordEmail sends a password reset email with a token.
func (s *Service) SendResetPasswordEmail(ctx context.Context, emailAddr, token string) error {
	resetURL := path.Join(s.config.BaseURL, "/reset-password?token=", token)

	subject := "Reset Your Password"
	body := fmt.Sprintf(`
Hello,

You requested to reset your password. Click the link below to reset it:

%s

This link will expire in 8 hours.

If you did not request this, please ignore this email.

Best regards,
Floxy Manager Team
`, resetURL)

	return s.sendEmail(ctx, emailAddr, subject, body)
}

// Send2FACodeEmail sends a 2FA code email for the specified action.
func (s *Service) Send2FACodeEmail(ctx context.Context, emailAddr, code, action string) error {
	var subject, body string

	switch action {
	case "disable":
		subject = "Disable Two-Factor Authentication"
		body = fmt.Sprintf(`
Hello,

You requested to disable two-factor authentication. Please use the following code to confirm:

%s

This code will expire in 15 minutes.

If you did not request this, please ignore this email and contact support immediately.

Best regards,
Floxy Manager Team
`, code)
	case "reset":
		subject = "Reset Two-Factor Authentication"
		body = fmt.Sprintf(`
Hello,

You requested to reset your two-factor authentication. Please use the following code to confirm:

%s

This code will expire in 15 minutes.

If you did not request this, please ignore this email and contact support immediately.

Best regards,
Floxy Manager Team
`, code)
	default:
		subject = "Two-Factor Authentication Code"
		body = fmt.Sprintf(`
Hello,

Your two-factor authentication code is:

%s

This code will expire in 15 minutes.

Best regards,
Floxy Manager Team
`, code)
	}

	return s.sendEmail(ctx, emailAddr, subject, body)
}

// sendEmail sends an email using SMTP.
func (s *Service) sendEmail(ctx context.Context, to, subject, body string) error {
	if s.config.SMTPHost == "" {
		slog.Warn("SMTP host is not configured, email sending is disabled")
		return nil
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, strings.Split(s.config.SMTPHost, ":")[0])

	// Format message
	msg := []byte(fmt.Sprintf("From: %s\r\n", s.config.From) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body)

	// Determine if we need TLS
	host := strings.Split(s.config.SMTPHost, ":")[0]
	addr := s.config.SMTPHost
	if !strings.Contains(addr, ":") {
		addr = host + ":25"
	}

	// Send email
	if s.config.UseTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: s.config.AllowInsecure,
		}

		if s.config.CertFile != "" && s.config.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
			if err != nil {
				return fmt.Errorf("load certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("create smtp client: %w", err)
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authenticate: %w", err)
		}

		if err := client.Mail(s.config.From); err != nil {
			return fmt.Errorf("set sender: %w", err)
		}

		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("set recipient: %w", err)
		}

		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("get data writer: %w", err)
		}

		if _, err := writer.Write(msg); err != nil {
			writer.Close()
			return fmt.Errorf("write message: %w", err)
		}

		if err := writer.Close(); err != nil {
			return fmt.Errorf("close writer: %w", err)
		}

		return client.Quit()
	}

	// Plain SMTP
	err := smtp.SendMail(addr, auth, s.config.From, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}
