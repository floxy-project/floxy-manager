package domain

import "time"

// ProductInfo represents product information including client ID.
type ProductInfo struct {
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}
