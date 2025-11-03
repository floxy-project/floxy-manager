package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rom8726/floxy-manager/internal/contract"
)

type SSOHandler struct {
	usersService contract.UsersUseCase
}

func NewSSOHandler(usersService contract.UsersUseCase) *SSOHandler {
	return &SSOHandler{
		usersService: usersService,
	}
}

func (h *SSOHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providers, err := h.usersService.GetSSOProviders(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get SSO providers")
		return
	}

	result := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		result = append(result, map[string]interface{}{
			"name":         p.GetName(),
			"display_name": p.GetDisplayName(),
			"icon_url":     p.GetIconURL(),
			"type":         p.GetType(),
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": result,
	})
}

func (h *SSOHandler) Initiate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProviderName string `json:"provider_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.ProviderName == "" {
		respondError(w, http.StatusBadRequest, "provider_name is required")
		return
	}

	redirectURL, err := h.usersService.SSOInitiate(r.Context(), req.ProviderName)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to initiate SSO")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"redirect_url": redirectURL,
	})
}

func (h *SSOHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	providerName := query.Get("provider")
	response := query.Get("SAMLResponse")
	state := query.Get("RelayState")

	if providerName == "" || response == "" {
		respondError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	accessToken, refreshToken, expiresIn, err := h.usersService.SSOCallback(r.Context(), providerName, r, response, state)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "SSO authentication failed")
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	frontendURL := scheme + "://" + r.Host
	redirectURL, _ := url.Parse(frontendURL)
	redirectURL.Path = "/auth/callback"
	q := redirectURL.Query()
	q.Set("access_token", accessToken)
	q.Set("refresh_token", refreshToken)
	q.Set("expires_in", fmt.Sprintf("%d", expiresIn))
	redirectURL.RawQuery = q.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

func (h *SSOHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		respondError(w, http.StatusBadRequest, "provider parameter is required")
		return
	}

	metadata, err := h.usersService.GetSSOMetadata(r.Context(), providerName)
	if err != nil {
		respondError(w, http.StatusNotFound, "Provider not found")
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write(metadata)
}
