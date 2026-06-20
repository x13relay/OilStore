package models

type Oil struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Visc  string `json:"visc"`
	Price int    `json:"price"`
}
