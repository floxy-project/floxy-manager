package domain

import (
	"time"
)

type LicenseType string

const (
	Trial           LicenseType = "trial"
	TrialSelfSigned LicenseType = "trial-self-signed"
	Commercial      LicenseType = "commercial"
)

type LicenseStatusType string

const (
	// Active represents an active license.
	Active LicenseStatusType = "active"
	// Expired represents an expired license.
	Expired LicenseStatusType = "expired"
	// Revoked represents a revoked license.
	Revoked LicenseStatusType = "revoked"
)

type License struct {
	ID          string      `json:"id"`
	ClientID    string      `json:"client_id"`
	Type        LicenseType `json:"type"`
	IssuedAt    time.Time   `json:"issued_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
	LicenseText string      `json:"license_text"`
	CreatedAt   time.Time   `json:"created_at"`
}

type LicenseStatus struct {
	ID              string      `json:"id"`
	Type            LicenseType `json:"type"`
	IssuedAt        time.Time   `json:"issued_at"`
	ExpiresAt       time.Time   `json:"expires_at"`
	IsValid         bool        `json:"is_valid"`
	IsExpired       bool        `json:"is_expired"`
	DaysUntilExpiry int         `json:"days_until_expiry"`
	LicenseText     string      `json:"license_text"`
}
