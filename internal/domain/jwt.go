package domain

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess        TokenType = "accessToken"
	TokenTypeRefresh       TokenType = "refreshToken"
	TokenTypeResetPassword TokenType = "resetPassword"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	TokenType   TokenType `json:"type"`
	UserID      uint      `json:"userId"`
	Username    string    `json:"username"`
	IsSuperuser bool      `json:"isSuperuser"`
}
