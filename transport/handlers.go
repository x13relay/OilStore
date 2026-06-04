package transport

import (
	dto "OilStore/DTO"
	"OilStore/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type Handlers struct {
	oilServ OilService
}

func NewHandlers(oilServ OilService) *Handlers {
	return &Handlers{
		oilServ: oilServ,
	}
}

func (h *Handlers) AddOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}
	var req dto.AddOilReq

	errJson := json.NewDecoder(r.Body).Decode(&req)
	if errJson != nil {
		http.Error(w, "Can't read jSon body", http.StatusBadRequest)
		return
	}
	newOil := models.Oil{
		Name:  req.Name,
		Visc:  req.Visc,
		Price: req.Price,
	}

	id, errDB := h.oilServ.AddOil(r.Context(), newOil)
	if errDB != nil {
		http.Error(w, "Fail to write DB", http.StatusInternalServerError)
		return
	}

	response := dto.OilMessageResp{
		Id:      id,
		Message: "Succes!New Oil created!",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) DeleteOilById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}
	errBD := h.oilServ.DeleteOilById(r.Context(), id)
	if errBD != nil {
		http.Error(w, "oil not found "+errBD.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	respMess := dto.OilMessageResp{
		Id:      id,
		Message: "Oil deleted successfully!",
	}
	json.NewEncoder(w).Encode(respMess)

}

func (h *Handlers) FullUpdateOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}
	var oilReq dto.FullUpdateOilReq

	idStr := r.PathValue("id") //парсим айди
	id, errId := strconv.Atoi(idStr)
	if errId != nil {
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}

	errJson := json.NewDecoder(r.Body).Decode(&oilReq)
	if errJson != nil {
		http.Error(w, "Can't read jSon body"+errJson.Error(), http.StatusBadRequest)
		return
	}

	updateOil := models.Oil{
		Name:  oilReq.Name,
		Visc:  oilReq.Visc,
		Price: oilReq.Price,
	}
	updOil, errBD := h.oilServ.FullUpdateOil(r.Context(), updateOil, id)
	if errBD != nil {
		if errors.Is(errBD, sql.ErrNoRows) {
			http.Error(w, "oil not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database write error: oil data not updated!"+errBD.Error(), http.StatusInternalServerError)
			return
		}
	}

	oilResp := dto.OilResp{
		Id:    updOil.Id,
		Name:  updOil.Name,
		Visc:  updOil.Visc,
		Price: updOil.Price,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	errJREQ := json.NewEncoder(w).Encode(oilResp)
	if errJREQ != nil {
		http.Error(w, "response error"+errJREQ.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handlers) GetMinMaxOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
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
	oils, errBD := h.oilServ.GetMinMaxOil(r.Context(), min, max)
	if errBD != nil {
		http.Error(w, "data base error!", http.StatusInternalServerError)
		return
	}

	oilRespList := dto.OilRespList{
		Data:  make([]dto.OilResp, len(oils)),
		Count: len(oils),
	}

	for i, oil := range oils {
		oilRespList.Data[i] = dto.OilResp{
			Id:    oil.Id,
			Name:  oil.Name,
			Visc:  oil.Visc,
			Price: oil.Price,
		}
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	if errJsonEnc := json.NewEncoder(w).Encode(oilRespList); errJsonEnc != nil {
		fmt.Println("Failed to encode response:", errJsonEnc)
	}

}

func (h *Handlers) GetByVisc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()

	visc := query.Get("visc")
	if visc == "" {
		http.Error(w, "empty viscosity!", http.StatusBadRequest)
		return
	}

	sortOil, err := h.oilServ.GetByVisc(r.Context(), visc)
	if err != nil {
		http.Error(w, "oil not found!", http.StatusInternalServerError)
		return
	}

	oilListResp := dto.OilRespList{
		Data:  make([]dto.OilResp, len(sortOil)),
		Count: len(sortOil),
	}

	for i, oil := range sortOil {
		oilListResp.Data[i] = dto.OilResp{
			Id:    oil.Id,
			Name:  oil.Name,
			Visc:  oil.Visc,
			Price: oil.Price,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if errJson := json.NewEncoder(w).Encode(oilListResp); errJson != nil {
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
		fmt.Println("Failed to encode response data")
		return
	}

}

func (h *Handlers) GetAllOils(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}

	allOil, err := h.oilServ.GetAllOils(r.Context())
	if err != nil {
		http.Error(w, "oils not found!", http.StatusInternalServerError)
		return
	}

	allOilList := dto.OilRespList{
		Data:  make([]dto.OilResp, len(allOil)),
		Count: len(allOil),
	}

	for i, oil := range allOil {
		allOilList.Data[i] = dto.OilResp{
			Id:    oil.Id,
			Name:  oil.Name,
			Visc:  oil.Visc,
			Price: oil.Price,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	errJson := json.NewEncoder(w).Encode(allOilList)
	if errJson != nil {
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetOilById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")

	id, errID := strconv.Atoi(idStr)
	if errID != nil {
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}
	resOil, errBD := h.oilServ.GetOilById(r.Context(), id)
	if errBD != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	oilResp := dto.OilResp{
		Id:    resOil.Id,
		Name:  resOil.Name,
		Visc:  resOil.Visc,
		Price: resOil.Price,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	errJson := json.NewEncoder(w).Encode(oilResp)
	if errJson != nil {
		http.Error(w, "Json error!"+errJson.Error(), http.StatusInternalServerError)
		return
	}
}
