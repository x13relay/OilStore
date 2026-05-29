package main

import (
	"OilStore/repository"
	"OilStore/service"
	"OilStore/transport"
	"context"
	"fmt"
	"net/http"
)

func main() {
	ctx := context.Background()
	conn, errCon := repository.ConnectionBD_oil(ctx)
	if errCon != nil {
		fmt.Println("Ошибка подключения к БД", errCon)
		return
	}
	defer conn.Close(ctx)

	oilRepo := repository.NewOilConn(conn)
	oilServ := service.NewOilService(oilRepo)
	handlers := transport.NewHandlers(oilServ)

	mux := http.NewServeMux()
	errBD := oilRepo.CreateTableOils(ctx)
	if errBD != nil {
		fmt.Println("Ошибка при создании таблицы в БД", errBD)
		return
	}
	fmt.Println("БД готова. Сервер запущен. Жду запросы на :8080")
	mux.HandleFunc("POST /oils", handlers.AddOil)
	mux.HandleFunc("DELETE /oils/{id}", handlers.DeleteOilById)
	mux.HandleFunc("PATCH /oils/{id}", handlers.FullUpdateOil)
	mux.HandleFunc("GET /oils", handlers.GetMinMaxOil)
	mux.HandleFunc("GET /oils/visc", handlers.GetByVisc)

	errServer := http.ListenAndServe(":8080", mux)
	if errServer != nil {
		fmt.Println("Ошибка сервера!", errServer)
		return
	}

}
