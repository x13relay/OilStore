package handlers

import (
	"OilStore/models"
	"OilStore/postgresql"
	"encoding/json"
	"net/http"
)

type Handlers struct {
	oilConn *postgresql.OilConn
}

func NewHandlers(oilConn *postgresql.OilConn) *Handlers {
	return &Handlers{
		oilConn: oilConn,
	}
}

func (h *Handlers) AddOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong HTTP method!", http.StatusBadRequest)
		return
	}
	var newOil models.Oil

	errJson := json.NewDecoder(r.Body).Decode(&newOil)
	if errJson != nil {
		http.Error(w, "Can't read jSon body", http.StatusBadRequest)
		return
	}
	id, errDB := h.oilConn.AddOil(r.Context(), newOil)
	if errDB != nil {
		http.Error(w, "Fail to write DB", 500)
		return
	}
	response := map[string]interface{}{
		"id":      id,
		"message": "Succes!New Oil created!",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
