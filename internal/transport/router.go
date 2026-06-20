package transport

import (
	"net/http"
)

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /oils", h.AddOil)
	mux.HandleFunc("DELETE /oils/{id}", h.DeleteOilById)
	mux.HandleFunc("PATCH /oils/{id}", h.FullUpdateOil)
	mux.HandleFunc("GET /oils/price", h.GetMinMaxOil)
	mux.HandleFunc("GET /oils/visc", h.GetByVisc)
	mux.HandleFunc("GET /oils", h.GetAllOils)
	mux.HandleFunc("GET /oils/pr/{price}", h.GetOilsAbovePrice)

	mux.HandleFunc("GET /oils/{id}", h.GetOilById)

	return mux
}
