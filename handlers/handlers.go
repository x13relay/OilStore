package handlers

import (
	"OilStore/models"
	"OilStore/postgresql"
	"encoding/json"
	"net/http"
	"strconv"
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

func (h *Handlers) DeleteOilById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Wrong HTTP method!", http.StatusBadRequest)
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}
	errBD := h.oilConn.DeleteOilById(r.Context(), id)
	if errBD != nil {
		http.Error(w, "oil not found"+errBD.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (h *Handlers) FullUpdateOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "wrong HTTP method!", http.StatusBadRequest)
		return
	}
	var updateOil models.Oil

	idStr := r.PathValue("id")
	id, errId := strconv.Atoi(idStr)
	if errId != nil {
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}

	errJson := json.NewDecoder(r.Body).Decode(&updateOil)
	if errJson != nil {
		http.Error(w, "Can't read jSon body"+errJson.Error(), http.StatusBadRequest)
		return
	}

	updOil, errBD := h.oilConn.FullUpdateOil(r.Context(), updateOil, id)
	if errBD != nil {
		http.Error(w, "Database write error: oil data not updated!"+errBD.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	errJREQ := json.NewEncoder(w).Encode(updOil)
	if errJREQ != nil {
		http.Error(w, "response error"+errJREQ.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handlers) GetMinMaxOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong HTTP method!", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()

	qMin := query.Get("min")
	qMax := query.Get("max")

	if qMin == "" || qMax == "" {
		http.Error(w, "incorrect minimum or maximum price", http.StatusBadRequest)
		return
	}

	min, errMin := strconv.Atoi(qMin)

	if errMin != nil {
		http.Error(w, "incorrect min price!", http.StatusBadRequest)
		return
	}

	max, errMax := strconv.Atoi(qMax)
	if errMax != nil {
		http.Error(w, "incorrect max price", http.StatusBadRequest)
		return
	}
	oils, errBD := h.oilConn.GetMinMaxOil(r.Context(), min, max)
	if errBD != nil {
		http.Error(w, "data base error!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(oils)
	w.WriteHeader(http.StatusOK)

}
