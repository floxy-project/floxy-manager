package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type SSOHandler struct {
	usersService contract.UsersUseCase
	frontendURL  string
}

func NewSSOHandler(usersService contract.UsersUseCase, frontendURL string) *SSOHandler {
	return &SSOHandler{
		usersService: usersService,
		frontendURL:  frontendURL,
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

// ACS handles SAML Assertion Consumer Service (ACS) endpoint.
// This endpoint receives POST requests from SAML Identity Providers with SAMLResponse and RelayState.
func (h *SSOHandler) ACS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	samlResponse := r.FormValue("SAMLResponse")
	relayState := r.FormValue("RelayState")

	slog.Info("SAML ACS endpoint called",
		"saml_response_length", len(samlResponse),
		"relay_state", relayState,
		"method", r.Method,
		"url", r.URL.String(),
	)

	if samlResponse == "" {
		respondError(w, http.StatusBadRequest, "Missing SAMLResponse parameter")
		return
	}

	// After ParseForm(), PostForm is already populated with form values
	// However, we need to ensure it's explicitly set for the SAML library
	ctx := r.Context()

	// Get or create raw request from context
	rawReq := appcontext.RawRequest(ctx)
	if rawReq == nil {
		rawReq = r
	}

	// Create a new request copy with PostForm explicitly set
	// This is needed because SAML library expects PostForm to be set
	rawReqCopy := rawReq.Clone(ctx)
	if rawReqCopy == nil {
		rawReqCopy = r
	}

	// Explicitly set PostForm for SAML library (as shown in user's example)
	if rawReqCopy.PostForm == nil {
		rawReqCopy.PostForm = make(url.Values)
	}
	rawReqCopy.PostForm.Set("SAMLResponse", samlResponse)
	if relayState != "" {
		rawReqCopy.PostForm.Set("RelayState", relayState)
	}

	// Put the request with PostForm in context for SAML processing
	ctx = appcontext.WithRawRequest(ctx, rawReqCopy)

	// Call SSOCallback with AD SAML provider
	accessToken, refreshToken, _, err := h.usersService.SSOCallback(
		ctx, domain.SSOProviderNameADSaml, rawReqCopy, samlResponse, relayState,
	)
	if err != nil {
		slog.Error("SSO assert failed", "error", err)
		if errors.Is(err, domain.ErrInactiveUser) {
			respondError(w, http.StatusUnauthorized, "user is inactive")

			return
		}

		respondError(w, http.StatusUnauthorized, "SSO authentication failed")

		return
	}

	// Build redirect URL to frontend
	redirectURL := h.buildFrontLoginSuccessLocation(accessToken, refreshToken)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// buildFrontLoginSuccessLocation builds the frontend success URL with tokens.
func (h *SSOHandler) buildFrontLoginSuccessLocation(accessToken, refreshToken string) string {
	values := url.Values{}
	values.Set("access_token", accessToken)
	values.Set("refresh_token", refreshToken)
	return h.frontendURL + "/auth/saml/success?" + values.Encode()
}
