package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
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

func (h *TenantsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can create tenants")
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	tenant, err := h.tenantsRepo.Create(r.Context(), req.Name)
	if err != nil {
		slog.Error("Failed to create tenant",
			"error", err,
			"name", req.Name,
		)
		respondError(w, http.StatusInternalServerError, "Failed to create tenant")
		return
	}

	respondJSON(w, http.StatusCreated, tenant)
}

func (h *TenantsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can delete tenants")
		return
	}

	// Get tenant ID from URL params
	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "tenant id is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	err = h.tenantsRepo.Delete(r.Context(), domain.TenantID(id))
	if err != nil {
		if err == domain.ErrEntityNotFound {
			respondError(w, http.StatusNotFound, "tenant not found")
			return
		}
		slog.Error("Failed to delete tenant",
			"error", err,
			"tenant_id", id,
		)
		respondError(w, http.StatusInternalServerError, "Failed to delete tenant")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "tenant deleted successfully"})
}
