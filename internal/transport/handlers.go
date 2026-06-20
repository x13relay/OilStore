package transport

import (
	"OilStore/internal/domain"
	dto "OilStore/internal/dto"
	"OilStore/internal/logger"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"
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
	var oilDomain domain.AddOilDomain

	errJson := json.NewDecoder(r.Body).Decode(&oilDomain)
	if errJson != nil {
		logger.Log.Error("Can't read request jSon body", zap.Error(errJson))
		http.Error(w, "Can't read jSon body", http.StatusBadRequest)
		return
	}

	id, errDB := h.oilServ.AddOil(r.Context(), oilDomain)
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
	errJsonResp := json.NewEncoder(w).Encode(response)
	if errJsonResp != nil {
		logger.Log.Error("Failed to encode JSON response", zap.Error(errJsonResp))
	}
}

func (h *Handlers) DeleteOilById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Log.Error("incorrect ID value!", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}

	errBD := h.oilServ.DeleteOilById(r.Context(), id)
	if errBD != nil {
		http.Error(w, "oil not found ", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	respMess := dto.OilMessageResp{
		Id:      id,
		Message: "Oil deleted successfully!",
	}
	errJsonResp := json.NewEncoder(w).Encode(respMess)
	if errJsonResp != nil {
		logger.Log.Error("Failed to encode JSON response", zap.Error(errJsonResp))
	}

}

func (h *Handlers) FullUpdateOil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}
	var oilReq dto.FullUpdateOilReq

	idStr := r.PathValue("id")
	id, errId := strconv.Atoi(idStr)
	if errId != nil {
		logger.Log.Error("incorrect ID value!", zap.String("id", idStr), zap.Error(errId))
		http.Error(w, "incorrect ID value!", http.StatusBadRequest)
		return
	}

	errJson := json.NewDecoder(r.Body).Decode(&oilReq)
	if errJson != nil {
		logger.Log.Error("Can't read jSon body", zap.Error(errJson))
		http.Error(w, "Can't read jSon body", http.StatusBadRequest)
		return
	}

	updateOil := domain.OilDomain{
		Name:  oilReq.Name,
		Visc:  oilReq.Visc,
		Price: oilReq.Price,
	}
	updOil, errBD := h.oilServ.FullUpdateOil(r.Context(), updateOil, id)
	if errBD != nil {
		if errors.Is(errBD, sql.ErrNoRows) {
			http.Error(w, "oil not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Database write error: oil data not updated!", http.StatusInternalServerError)
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
	errJsonResp := json.NewEncoder(w).Encode(oilResp)
	if errJsonResp != nil {
		logger.Log.Error("Failed to encode JSON response", zap.Error(errJsonResp))
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
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
		logger.Log.Error("minimum and maximum price can not be empty", zap.String("min", qMin), zap.String("max", qMax))
		http.Error(w, " minimum and maximum price can not be empty", http.StatusBadRequest)
		return
	}

	min, errMin := strconv.Atoi(qMin)

	if errMin != nil {
		logger.Log.Error("incorrect min price!", zap.Error(errMin), zap.String("min", qMin))
		http.Error(w, "incorrect min price!", http.StatusBadRequest)
		return
	}

	max, errMax := strconv.Atoi(qMax)
	if errMax != nil {
		logger.Log.Error("incorrect max price!", zap.Error(errMax), zap.String("max", qMax))
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
	if errJsonResp := json.NewEncoder(w).Encode(oilRespList); errJsonResp != nil {

		logger.Log.Error("Failed to encode JSON response", zap.Error(errJsonResp))
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
		logger.Log.Error("The field viscosity cannot be empty !")
		http.Error(w, "The field viscosity cannot be empty !", http.StatusBadRequest)
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
	if errJsonResp := json.NewEncoder(w).Encode(oilListResp); errJsonResp != nil {
		logger.Log.Error("Failed to encode response data", zap.Error(errJsonResp))
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
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
	errJsonResp := json.NewEncoder(w).Encode(allOilList)
	if errJsonResp != nil {
		logger.Log.Error("Failed to encode response data", zap.Error(errJsonResp))
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
		logger.Log.Error("incorrect ID value!", zap.String("id", idStr), zap.Error(errID))
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
	errJsonResp := json.NewEncoder(w).Encode(oilResp)
	if errJsonResp != nil {
		logger.Log.Error("Failed to encode response data", zap.Error(errJsonResp))
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetOilsAbovePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong HTTP method!", http.StatusMethodNotAllowed)
		return
	}
	priceStr := r.PathValue("price")
	price, errPrice := strconv.Atoi(priceStr)
	if errPrice != nil {
		logger.Log.Error("Wrong price value!", zap.String("price", priceStr), zap.Error(errPrice))
		http.Error(w, "Wrong price value!", http.StatusBadRequest)
		return
	}

	newOilsSlice, errService := h.oilServ.GetOilsAbovePrice(r.Context(), price)
	if errService != nil {
		http.Error(w, "error!", http.StatusInternalServerError)
		return
	}

	newOils := dto.OilRespList{
		Data:  make([]dto.OilResp, len(newOilsSlice)),
		Count: len(newOilsSlice),
	}

	for i, oil := range newOilsSlice {
		newOils.Data[i] = dto.OilResp{
			Id:    oil.Id,
			Name:  oil.Name,
			Visc:  oil.Visc,
			Price: oil.Price,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	errJsonResp := json.NewEncoder(w).Encode(newOils)
	if errJsonResp != nil {
		logger.Log.Error("Failed to encode response data", zap.Error(errJsonResp))
		http.Error(w, "Failed to encode response data", http.StatusInternalServerError)
		return
	}

}
