package dto

type OilResp struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Visc  string `json:"visc"`
	Price int    `json:"price"`
}

type OilRespList struct {
	Data  []OilResp `json:"data"`
	Count int       `json:"count"`
}

type OilMessageResp struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}
