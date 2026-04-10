package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"currencyparser/backend/internal/service"
)

type ctxKey string

const userIDContextKey ctxKey = "userID"

type Handler struct {
	authService  *service.AuthService
	ratesService *service.RatesService
	frontendHost string
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(authService *service.AuthService, ratesService *service.RatesService, frontendHost string) *Handler {
	return &Handler{
		authService:  authService,
		ratesService: ratesService,
		frontendHost: frontendHost,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("POST /api/v1/auth/register", h.handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.requireAuth(h.handleLogout))
	mux.HandleFunc("GET /api/v1/auth/me", h.requireAuth(h.handleMe))
	mux.HandleFunc("GET /api/v1/teams", h.requireAuth(h.handleCurrencies))
	mux.HandleFunc("GET /api/v1/matchups/index", h.requireAuth(h.handleConvert))
	mux.HandleFunc("GET /api/v1/matchups/live", h.requireAuth(h.handleLatest))
	mux.HandleFunc("GET /api/v1/matchups/history", h.requireAuth(h.handleHistory))
	mux.HandleFunc("GET /api/v1/esports/digest", h.requireAuth(h.handleEsportsDigest))
}

func (h *Handler) handleEsportsDigest(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, service.EsportsDigest())
}

func (h *Handler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		allowedOrigin := resolveAllowedOrigin(h.frontendHost, origin)
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func resolveAllowedOrigin(frontendHost, requestOrigin string) string {
	allowed := strings.TrimSpace(frontendHost)
	if allowed == "" {
		return ""
	}
	if allowed == "*" {
		return "*"
	}

	parts := strings.Split(allowed, ",")
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		if requestOrigin == "" || strings.EqualFold(origin, requestOrigin) {
			return origin
		}
	}

	return ""
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		userID, err := h.authService.ParseAccessToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid access token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	pair, err := h.authService.Register(r.Context(), request.Email, request.Password, request.Username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, pair)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	pair, err := h.authService.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	pair, err := h.authService.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.authService.Logout(r.Context(), request.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.CurrentUser(r.Context(), userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleCurrencies(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": h.ratesService.ListCurrencies()})
}

func (h *Handler) handleConvert(w http.ResponseWriter, r *http.Request) {
	amount, err := parseFloatParamWithFallback(r, "weight", "amount")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	team := queryWithFallback(r, "team", "base")
	opponent := queryWithFallback(r, "opponent", "target")
	response, err := h.ratesService.Convert(r.Context(), team, opponent, amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleLatest(w http.ResponseWriter, r *http.Request) {
	symbolsRaw := queryWithFallback(r, "opponents", "symbols")
	symbols := strings.Split(strings.TrimSpace(symbolsRaw), ",")
	response, err := h.ratesService.Latest(r.Context(), queryWithFallback(r, "team", "base"), symbols)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from date")
		return
	}
	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to date")
		return
	}
	response, err := h.ratesService.History(
		r.Context(),
		queryWithFallback(r, "team", "base"),
		queryWithFallback(r, "opponent", "target"),
		from,
		to,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func parseFloatParam(r *http.Request, name string) (float64, error) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, errors.New(name + " is required")
	}
	return strconv.ParseFloat(value, 64)
}

func parseFloatParamWithFallback(r *http.Request, primary, fallback string) (float64, error) {
	value := strings.TrimSpace(r.URL.Query().Get(primary))
	if value == "" && fallback != "" {
		value = strings.TrimSpace(r.URL.Query().Get(fallback))
	}
	if value == "" {
		return 0, errors.New(primary + " is required")
	}
	return strconv.ParseFloat(value, 64)
}

func queryWithFallback(r *http.Request, primary, fallback string) string {
	value := strings.TrimSpace(r.URL.Query().Get(primary))
	if value != "" {
		return value
	}
	return strings.TrimSpace(r.URL.Query().Get(fallback))
}

func userIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDContextKey).(string)
	return userID
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
