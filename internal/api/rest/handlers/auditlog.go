package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type AuditLogHandler struct {
	auditLogRepo   contract.AuditLogRepository
	permissionsSrv contract.PermissionsService
}

func NewAuditLogHandler(
	auditLogRepo contract.AuditLogRepository,
	permissionsSrv contract.PermissionsService,
) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogRepo:   auditLogRepo,
		permissionsSrv: permissionsSrv,
	}
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	projectIDInt, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}
	projectID := domain.ProjectID(projectIDInt)

	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	if !appcontext.IsSuper(r.Context()) {
		if err := h.permissionsSrv.CanViewAudit(r.Context(), projectID); err != nil {
			if errors.Is(err, domain.ErrPermissionDenied) {
				respondError(w, http.StatusForbidden, "Access denied to audit log")
				return
			}
			respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
			return
		}
	}

	page, pageSize := parsePagination(r)

	entries, total, err := h.auditLogRepo.List(r.Context(), projectID, page, pageSize)
	if err != nil {
		slog.Error("Failed to list audit log",
			"error", err,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     entries,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}
