package handlers

import (
	"encoding/json"
	"net/http"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
)

type TwoFAHandler struct {
	usersService contract.UsersUseCase
}

func NewTwoFAHandler(usersService contract.UsersUseCase) *TwoFAHandler {
	return &TwoFAHandler{
		usersService: usersService,
	}
}

func (h *TwoFAHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code      string `json:"code"`
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.SessionID == "" {
		respondError(w, http.StatusBadRequest, "code and session_id are required")
		return
	}

	accessToken, refreshToken, expiresIn, err := h.usersService.Verify2FA(r.Context(), req.Code, req.SessionID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid 2FA code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    expiresIn,
	})
}

func (h *TwoFAHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	secret, qrURL, qrImage, err := h.usersService.Setup2FA(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to setup 2FA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"secret":   secret,
		"qr_url":   qrURL,
		"qr_image": qrImage,
	})
}

func (h *TwoFAHandler) Confirm2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		respondError(w, http.StatusBadRequest, "code is required")
		return
	}

	err := h.usersService.Confirm2FA(r.Context(), userID, req.Code)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "2FA enabled successfully",
	})
}

func (h *TwoFAHandler) Send2FACode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.usersService.Send2FACode(r.Context(), userID, req.Action)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "2FA code sent to your email",
	})
}

func (h *TwoFAHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		EmailCode string `json:"email_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.EmailCode == "" {
		respondError(w, http.StatusBadRequest, "email_code is required")
		return
	}

	err := h.usersService.Disable2FA(r.Context(), userID, req.EmailCode)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid email code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "2FA disabled successfully",
	})
}

func (h *TwoFAHandler) Reset2FA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		EmailCode string `json:"email_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.EmailCode == "" {
		respondError(w, http.StatusBadRequest, "email_code is required")
		return
	}

	secret, qrURL, qrImage, err := h.usersService.Reset2FA(r.Context(), userID, req.EmailCode)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid email code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"secret":   secret,
		"qr_url":   qrURL,
		"qr_image": qrImage,
	})
}
