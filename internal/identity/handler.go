package identity

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"parkwatch/pkg/httpx"
)

// Handler exposes the identity HTTP API. It sits behind the gateway, so it does
// not deal with cookies or CORS — only request/response bodies.
type Handler struct{ svc *Service }

// NewHandler builds the handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes returns the identity router.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", httpx.Health)
	r.Post("/login", h.login)
	return r
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		httpx.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	res, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"token": res.Token,
		"user":  res.User,
	})
}
