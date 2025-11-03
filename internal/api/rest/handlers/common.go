package handlers

import (
	"net/http"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
)

// requireAuth checks if the user is authenticated in the context.
// Returns true if authenticated, false otherwise.
func requireAuth(r *http.Request) bool {
	userID := appcontext.UserID(r.Context())
	return userID != 0
}

// checkAuthAndRespond checks authentication and responds with 401 if not authenticated.
// Returns true if authenticated, false otherwise.
func checkAuthAndRespond(w http.ResponseWriter, r *http.Request) bool {
	if !requireAuth(r) {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return false
	}
	return true
}
