package contract

import (
	"github.com/rom8726/floxy-manager/internal/domain"
)

type TwoFARateLimiter interface {
	Inc(userID domain.UserID) (attempts int, blocked bool)
	Reset(userID domain.UserID)
	IsBlocked(userID domain.UserID) bool
}
