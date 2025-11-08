package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/livekit/livekit-server/pkg/auth"
	"github.com/livekit/livekit-server/pkg/storage"
)

const defaultTokenTTL = 24 * time.Hour

type authRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type authUserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type authResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

type AuthHandler struct {
	service *auth.Service
	tokens  *auth.TokenGenerator
}

func NewAuthHandler(service *auth.Service, tokens *auth.TokenGenerator) *AuthHandler {
	return &AuthHandler{service: service, tokens: tokens}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		// Check for duplicate user error (UNIQUE constraint violation)
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "unique") || 
		   strings.Contains(errMsg, "duplicate") || 
		   strings.Contains(errMsg, "already exists") ||
		   strings.Contains(errMsg, "23505") { // PostgreSQL unique violation code
			h.writeError(w, "email already registered", http.StatusConflict)
			return
		}
		h.writeError(w, "failed to register: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeAuthResponse(w, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.writeError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		h.writeError(w, "failed to login: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeAuthResponse(w, user)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (h *AuthHandler) writeAuthResponse(w http.ResponseWriter, user *storage.User) {
	token, err := h.tokens.Generate(user.ID, defaultTokenTTL)
	if err != nil {
		h.writeError(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	resp := authResponse{
		Token: token,
		User: authUserResponse{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName.String,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
