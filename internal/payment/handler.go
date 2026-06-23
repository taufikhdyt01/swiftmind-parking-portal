package payment

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"parkwatch/pkg/domain"
	"parkwatch/pkg/httpx"
)

// Handler exposes the payment HTTP API behind the gateway.
type Handler struct{ svc *Service }

// NewHandler builds the handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes returns the payment router. RBAC (members pay) is enforced at the gateway.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", httpx.Health)
	r.Get("/", h.list)
	r.Post("/{id}/pay", h.pay)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.svc.List(r.Context(),
		r.Header.Get(domain.HeaderUserRole), r.Header.Get(domain.HeaderUserEmail))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if invoices == nil {
		invoices = []Invoice{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"invoices": invoices})
}

type payRequest struct {
	Scenario string `json:"scenario"`
}

func (h *Handler) pay(w http.ResponseWriter, r *http.Request) {
	var req payRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Pay(r.Context(), chi.URLParam(r, "id"), req.Scenario,
		r.Header.Get(domain.HeaderUserEmail))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httpx.Error(w, http.StatusNotFound, "invoice not found")
		case errors.Is(err, ErrForbidden):
			httpx.Error(w, http.StatusForbidden, "not your invoice")
		case errors.Is(err, ErrInvalidScenario):
			httpx.Error(w, http.StatusBadRequest, "scenario must be success or failed")
		case errors.Is(err, ErrBusy):
			httpx.Error(w, http.StatusConflict, "a payment for this invoice is already in progress")
		default:
			httpx.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	httpx.JSON(w, http.StatusOK, result)
}
