package contract

import (
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type Tokenizer interface {
	AccessToken(user *domain.User) (string, error)
	RefreshToken(user *domain.User) (string, error)
	VerifyToken(token string, tokenType domain.TokenType) (*domain.TokenClaims, error)
	ResetPasswordToken(user *domain.User) (string, time.Duration, error)
	AccessTokenTTL() time.Duration
	SecretKey() string
}
