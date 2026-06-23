package rules

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"parkwatch/pkg/domain"
	"parkwatch/pkg/fine"
	"parkwatch/pkg/httpx"
)

// Handler exposes the rules HTTP API. It runs behind the gateway and trusts the
// X-User-* headers the gateway sets after verifying the JWT.
type Handler struct{ svc *Service }

// NewHandler builds the handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes returns the rules router. RBAC (officer-only publish) is enforced at
// the gateway; this service trusts what reaches it.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", httpx.Health)
	r.Get("/active", h.active)
	r.Get("/", h.list)
	r.Post("/", h.publish)
	return r
}

func (h *Handler) active(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Active(r.Context())
	if err != nil {
		if errors.Is(err, ErrNoActiveRuleset) {
			httpx.Error(w, http.StatusNotFound, "no active ruleset")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.JSON(w, http.StatusOK, v)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	versions, err := h.svc.List(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if versions == nil {
		versions = []Version{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"versions": versions})
}

func (h *Handler) publish(w http.ResponseWriter, r *http.Request) {
	var ruleset fine.Ruleset
	if err := json.NewDecoder(r.Body).Decode(&ruleset); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdBy := r.Header.Get(domain.HeaderUserEmail)
	if createdBy == "" {
		createdBy = "unknown"
	}

	v, err := h.svc.Publish(r.Context(), ruleset, createdBy)
	if err != nil {
		// Ruleset validation errors are user-facing; everything else is internal.
		if errors.Is(err, fine.ErrInvalidRuleset) {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.JSON(w, http.StatusCreated, v)
}
