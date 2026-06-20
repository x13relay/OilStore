package dto

type AddOilReq struct {
	Name  string `json:"name"`
	Visc  string `json:"visc"`
	Price int    `json:"price"`
}

type FullUpdateOilReq struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Visc  string `json:"visc"`
	Price int    `json:"price"`
}
