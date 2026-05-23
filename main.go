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
	fmt.Println("БД готова")
	mux.HandleFunc("/oils", handlers.AddOil)
	fmt.Println("Сервер запущен. Жду запросы на :8080")
	errServer := http.ListenAndServe(":8080", mux)
	if errServer != nil {
		fmt.Println("Ошибка сервера!", errServer)
		return
	}

}
