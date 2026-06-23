package violation

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"parkwatch/pkg/domain"
	"parkwatch/pkg/httpx"
)

const maxPhotoBytes = 10 << 20 // 10 MiB

// Handler exposes the violation HTTP API behind the gateway.
type Handler struct{ svc *Service }

// NewHandler builds the handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes returns the violation router. RBAC (officers create) is at the gateway.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", httpx.Health)
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}", h.get)
	r.Get("/{id}/photo", h.photo)
	return r
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"),
		r.Header.Get(domain.HeaderUserRole), r.Header.Get(domain.HeaderUserEmail))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "violation not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.JSON(w, http.StatusOK, v)
}

func (h *Handler) photo(w http.ResponseWriter, r *http.Request) {
	rc, contentType, err := h.svc.Photo(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "photo not found")
		return
	}
	defer rc.Close()
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_, _ = io.Copy(w, rc)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxPhotoBytes); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	occurredAt, err := parseTimestamp(r.FormValue("occurred_at"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid occurred_at timestamp")
		return
	}

	in := CreateInput{
		Plate:         r.FormValue("plate"),
		ViolationType: r.FormValue("violation_type"),
		Location:      r.FormValue("location"),
		OccurredAt:    occurredAt,
		IssuedByEmail: r.Header.Get(domain.HeaderUserEmail),
	}

	var photo *Photo
	file, header, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		photo = &Photo{
			Reader:      file,
			Size:        header.Size,
			ContentType: header.Header.Get("Content-Type"),
			Filename:    header.Filename,
		}
	}

	v, err := h.svc.Create(r.Context(), in, photo)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httpx.Error(w, http.StatusBadRequest, "invalid violation details")
			return
		}
		httpx.Error(w, http.StatusBadGateway, "could not create violation")
		return
	}
	httpx.JSON(w, http.StatusCreated, v)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	violations, err := h.svc.List(r.Context(),
		r.Header.Get(domain.HeaderUserRole), r.Header.Get(domain.HeaderUserEmail))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if violations == nil {
		violations = []Violation{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"violations": violations})
}

// parseTimestamp accepts RFC3339 or the HTML datetime-local format.
func parseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unrecognized timestamp")
}
