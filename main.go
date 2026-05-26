package main

import (
	"OilStore/handlers"
	"OilStore/postgresql"
	"context"
	"fmt"
	"net/http"
)

func main() {
	ctx := context.Background()
	conn, errCon := postgresql.ConnectionBD_oil(ctx)
	if errCon != nil {
		fmt.Println("Ошибка подключения к БД", errCon)
		return
	}
	oilConn := postgresql.NewOilConn(conn)
	handlers := handlers.NewHandlers(oilConn)
	mux := http.NewServeMux()
	errBD := oilConn.CreateTableOils(ctx)
	if errBD != nil {
		fmt.Println("Ошибка при создании таблицы в БД", errBD)
		return
	}
	fmt.Println("БД готова. Сервер запущен. Жду запросы на :8080")
	mux.HandleFunc("POST /oils", handlers.AddOil)
	mux.HandleFunc("DELETE /del/{id}", handlers.DeleteOilById)
	mux.HandleFunc("PATCH /oils/{id}", handlers.FullUpdateOil)
	mux.HandleFunc("GET /oils", handlers.GetMinMaxOil)

	errServer := http.ListenAndServe(":8080", mux)
	if errServer != nil {
		fmt.Println("Ошибка сервера!", errServer)
		return
	}

}
