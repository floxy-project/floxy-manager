package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type AuthHandler struct {
	usersService contract.UsersUseCase
}

func NewAuthHandler(usersService contract.UsersUseCase) *AuthHandler {
	return &AuthHandler{
		usersService: usersService,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UsernameOrEmail string `json:"username_or_email"`
		Password        string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.UsernameOrEmail == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "username_or_email and password are required")
		return
	}

	accessToken, refreshToken, sessionID, isTmpPassword, err := h.usersService.Login(r.Context(), req.UsernameOrEmail, req.Password)
	if err != nil {
		if err == domain.ErrTwoFARequired {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"session_id":   sessionID,
				"requires_2fa": true,
			})
			return
		}
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":    accessToken,
		"refresh_token":   refreshToken,
		"is_tmp_password": isTmpPassword,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	accessToken, refreshToken, err := h.usersService.LoginReissue(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
