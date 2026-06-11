package main

import (
	"OilStore/rdb"
	"OilStore/repository"
	"OilStore/service"
	"OilStore/transport"
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	rdb := rdb.RedisInit()
	fmt.Println("✅ Redis готов")
	defer rdb.Close()
	ctx := context.Background()
	conn, errCon := repository.ConnectionBD_oil(ctx) //создаем подключение к БД
	if errCon != nil {
		fmt.Println("Ошибка подключения к БД", errCon)
		return
	}
	fmt.Println("✅ PostgreSQL готов")
	defer conn.Close(ctx)

	oilRepo := repository.NewOilConn(conn)
	oilServ := service.NewOilService(oilRepo, rdb)
	handlers := transport.NewHandlers(oilServ)

	mux := http.NewServeMux()
	errBD := oilRepo.CreateTableOils(ctx)
	if errBD != nil {
		fmt.Println("Ошибка при создании таблицы в БД", errBD)
		return
	}

	mux.HandleFunc("POST /oils", handlers.AddOil)
	mux.HandleFunc("DELETE /oils/{id}", handlers.DeleteOilById)
	mux.HandleFunc("PATCH /oils/{id}", handlers.FullUpdateOil)
	mux.HandleFunc("GET /oils/price", handlers.GetMinMaxOil)
	mux.HandleFunc("GET /oils/visc", handlers.GetByVisc)
	mux.HandleFunc("GET /oils", handlers.GetAllOils)
	mux.HandleFunc("GET /oils/pr/{price}", handlers.GetOilsAbovePrice)

	mux.HandleFunc("GET /oils/{id}", handlers.GetOilById)

	ctxStop, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	oilSrv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("✅ Сервер запущен. Жду запросы на :8080")
		if errServer := oilSrv.ListenAndServe(); errServer != nil && errServer != http.ErrServerClosed {
			fmt.Println("Server Error!", errServer)
		}
	}()

	<-ctxStop.Done()
	fmt.Println("🛑 Получен сигнал остановки. Завершаю работу приложения и сервера...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := oilSrv.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Ошибка остановки сервера...")
	}
	fmt.Println("⛔ Сервер завершил свою работу.")
}
