package handlers

import (
	"OilStore/postgresql"
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

}
