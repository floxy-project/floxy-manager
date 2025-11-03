package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type PasswordHandler struct {
	usersService contract.UsersUseCase
}

func NewPasswordHandler(usersService contract.UsersUseCase) *PasswordHandler {
	return &PasswordHandler{
		usersService: usersService,
	}
}

func (h *PasswordHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	err := h.usersService.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

func (h *PasswordHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "token and new_password are required")
		return
	}

	err := h.usersService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

func (h *PasswordHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "new_password is required")
		return
	}

	if len(req.NewPassword) < 6 {
		respondError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	err := h.usersService.ChangeTemporaryPassword(r.Context(), userID, req.NewPassword)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "You don't have a temporary password or are not allowed to change it")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
