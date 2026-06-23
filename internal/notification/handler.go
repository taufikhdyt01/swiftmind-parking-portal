package notification

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"parkwatch/pkg/domain"
	"parkwatch/pkg/httpx"
)

// Handler exposes the notification HTTP API behind the gateway.
type Handler struct{ svc *Service }

// NewHandler builds the handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes returns the notification router.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", httpx.Health)
	r.Get("/", h.list)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	notifications, err := h.svc.List(r.Context(), r.Header.Get(domain.HeaderUserEmail))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if notifications == nil {
		notifications = []Notification{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"notifications": notifications})
}
