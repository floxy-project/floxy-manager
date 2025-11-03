package licenses

import (
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type licenseModel struct {
	ID          string    `db:"id"`
	LicenseText string    `db:"license_text"`
	IssuedAt    time.Time `db:"issued_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	ClientID    string    `db:"client_id"`
	Type        string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`
}

func (m *licenseModel) toDomain() domain.License {
	return domain.License{
		ID:          m.ID,
		ClientID:    m.ClientID,
		Type:        domain.LicenseType(m.Type),
		IssuedAt:    m.IssuedAt,
		ExpiresAt:   m.ExpiresAt,
		LicenseText: m.LicenseText,
		CreatedAt:   m.CreatedAt,
	}
}

func fromDomain(license domain.License) licenseModel {
	return licenseModel{
		ID:          license.ID,
		LicenseText: license.LicenseText,
		IssuedAt:    license.IssuedAt,
		ExpiresAt:   license.ExpiresAt,
		ClientID:    license.ClientID,
		Type:        string(license.Type),
		CreatedAt:   license.CreatedAt,
	}
}
