package domain

type AddOilDomain struct {
	Name  string `json:"name"`
	Visc  string `json:"visc"`
	Price int    `json:"price"`
}

type OilDomain struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Visc  string `json:"visc"`
	Price int    `json:"price"`
}
