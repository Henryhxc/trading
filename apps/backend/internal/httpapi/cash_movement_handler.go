package httpapi

import (
	"net/http"
	"ux-securities/backend/internal/domain/cashmovement"
)

func (s *server) cashMovements(w http.ResponseWriter, r *http.Request) {
	items, err := s.cash.List(r.Context(), actorFrom(r))
	if err != nil {
		s.fail(w, r, http.StatusInternalServerError, "internal_error", "Cash mock data is unavailable")
		return
	}
	if items == nil { items = []cashmovement.Movement{} }
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
