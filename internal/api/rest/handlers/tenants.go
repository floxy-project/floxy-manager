package handlers

import (
	"log/slog"
	"net/http"

	"github.com/rom8726/floxy-manager/internal/contract"
)

type TenantsHandler struct {
	tenantsRepo contract.TenantsRepository
}

func NewTenantsHandler(tenantsRepo contract.TenantsRepository) *TenantsHandler {
	return &TenantsHandler{
		tenantsRepo: tenantsRepo,
	}
}

func (h *TenantsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	tenants, err := h.tenantsRepo.List(r.Context())
	if err != nil {
		slog.Error("Failed to list tenants",
			"error", err,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tenants)
}
