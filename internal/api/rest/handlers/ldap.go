package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type LDAPHandler struct {
	ldapUseCase     contract.LDAPSyncUseCase
	settingsUseCase contract.SettingsUseCase
}

func NewLDAPHandler(ldapUseCase contract.LDAPSyncUseCase, settingsUseCase contract.SettingsUseCase) *LDAPHandler {
	return &LDAPHandler{
		ldapUseCase:     ldapUseCase,
		settingsUseCase: settingsUseCase,
	}
}

// GetLDAPConfig returns the current LDAP configuration
func (h *LDAPHandler) GetLDAPConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can access LDAP configuration")
		return
	}

	ctx := r.Context()
	config, err := h.settingsUseCase.GetLDAPConfig(ctx)
	if err != nil {
		respondError(w, http.StatusNotFound, "LDAP configuration not found")
		return
	}

	respondJSON(w, http.StatusOK, config)
}

// UpdateLDAPConfig updates the LDAP configuration
func (h *LDAPHandler) UpdateLDAPConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can update LDAP configuration")
		return
	}

	var config domain.LDAPConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	if err := h.ldapUseCase.UpdateConfig(ctx, &config); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update LDAP configuration")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "LDAP configuration updated successfully",
		"config":  config,
	})
}

// DeleteLDAPConfig deletes the LDAP configuration (disables LDAP)
func (h *LDAPHandler) DeleteLDAPConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can delete LDAP configuration")
		return
	}

	ctx := r.Context()
	if err := h.ldapUseCase.Disable(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to disable LDAP")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "LDAP configuration deleted successfully",
	})
}

// TestLDAPConnection tests the connection to the LDAP server
func (h *LDAPHandler) TestLDAPConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can test LDAP connection")
		return
	}

	var testConfig domain.LDAPConfig
	if err := json.NewDecoder(r.Body).Decode(&testConfig); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()

	// Ensure enabled is set to true for testing
	testConfig.Enabled = true

	// Save current config if exists
	var oldConfig *domain.LDAPConfig
	if savedConfig, err := h.settingsUseCase.GetLDAPConfig(ctx); err == nil {
		oldConfig = savedConfig
	}

	// Update to test config temporarily
	if err := h.ldapUseCase.UpdateConfig(ctx, &testConfig); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to update config for test")
		return
	}

	// Force reload config synchronously before testing
	// UpdateConfig triggers ReloadConfig asynchronously, so we need to reload manually
	// to ensure the test uses the updated configuration
	if err := h.ldapUseCase.ReloadConfig(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reload config for test")
		return
	}

	// Test connection
	err := h.ldapUseCase.TestConnection(ctx)

	// Restore old config if it existed
	if oldConfig != nil {
		_ = h.ldapUseCase.UpdateConfig(ctx, oldConfig)
	} else {
		// If no old config, disable LDAP
		_ = h.ldapUseCase.Disable(ctx)
	}

	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Connection test successful",
	})
}

// SyncLDAPUsers starts a manual LDAP user synchronization
func (h *LDAPHandler) SyncLDAPUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can start LDAP synchronization")
		return
	}

	ctx := r.Context()
	if err := h.ldapUseCase.StartManualSync(ctx); err != nil {
		if err.Error() == "a sync is already running" {
			respondError(w, http.StatusConflict, "Sync already in progress")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to start synchronization")
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":            "Synchronization started",
		"sync_id":            "", // Will be filled by progress endpoint
		"estimated_duration": "5m",
	})
}

// CancelLDAPSync cancels an ongoing LDAP synchronization
func (h *LDAPHandler) CancelLDAPSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can cancel LDAP synchronization")
		return
	}

	ctx := r.Context()
	if err := h.ldapUseCase.CancelSync(ctx); err != nil {
		if err.Error() == "no sync is currently running" {
			respondError(w, http.StatusNotFound, "No active synchronization")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to cancel synchronization")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Synchronization cancelled successfully",
	})
}

// GetLDAPSyncStatus returns the current LDAP synchronization status
func (h *LDAPHandler) GetLDAPSyncStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can access LDAP sync status")
		return
	}

	ctx := r.Context()
	status, err := h.ldapUseCase.GetSyncStatus(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sync status")
		return
	}

	// Convert to JSON response format
	response := map[string]interface{}{
		"status":             status.Status,
		"is_running":         status.IsRunning,
		"total_users":        status.TotalUsers,
		"synced_users":       status.SyncedUsers,
		"errors":             status.Errors,
		"warnings":           status.Warnings,
		"last_sync_duration": status.LastSyncDuration,
	}

	if !status.LastSyncTime.IsZero() {
		response["last_sync_time"] = status.LastSyncTime.Format(time.RFC3339)
	}

	respondJSON(w, http.StatusOK, response)
}

// GetLDAPSyncProgress returns the progress of an ongoing LDAP synchronization
func (h *LDAPHandler) GetLDAPSyncProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can access LDAP sync progress")
		return
	}

	ctx := r.Context()
	progress, err := h.ldapUseCase.GetSyncProgress(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sync progress")
		return
	}

	response := map[string]interface{}{
		"is_running":      progress.IsRunning,
		"progress":        progress.Progress,
		"processed_items": progress.ProcessedItems,
		"total_items":     progress.TotalItems,
		"sync_id":         progress.SyncID,
	}

	if progress.CurrentStep != "" {
		response["current_step"] = progress.CurrentStep
	}
	if progress.EstimatedTime != "" {
		response["estimated_time"] = progress.EstimatedTime
	}
	if !progress.StartTime.IsZero() {
		response["start_time"] = progress.StartTime.Format(time.RFC3339)
	}

	respondJSON(w, http.StatusOK, response)
}

// GetLDAPSyncLogs returns LDAP synchronization logs with filtering
func (h *LDAPHandler) GetLDAPSyncLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can access LDAP sync logs")
		return
	}

	// Parse query parameters
	filter := domain.LDAPSyncLogFilter{}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = &limit
		}
	}

	if level := r.URL.Query().Get("level"); level != "" {
		filter.Level = &level
	}

	if syncId := r.URL.Query().Get("sync_id"); syncId != "" {
		filter.SyncID = &syncId
	}

	if username := r.URL.Query().Get("username"); username != "" {
		filter.Username = &username
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &from
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &to
		}
	}

	ctx := r.Context()
	result, err := h.ldapUseCase.GetSyncLogs(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sync logs")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetLDAPSyncLogDetails returns details of a specific sync log entry
func (h *LDAPHandler) GetLDAPSyncLogDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can access LDAP sync log details")
		return
	}

	// Get log ID from URL path
	ps := appcontext.Params(r.Context())
	idStr := ps.ByName("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Log ID is required")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid log ID")
		return
	}

	ctx := r.Context()
	log, err := h.ldapUseCase.GetSyncLogDetails(ctx, uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "Log not found")
		return
	}

	respondJSON(w, http.StatusOK, log)
}

// GetLDAPStatistics returns comprehensive LDAP statistics
func (h *LDAPHandler) GetLDAPStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can access LDAP statistics")
		return
	}

	ctx := r.Context()
	stats, err := h.ldapUseCase.GetStatistics(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
