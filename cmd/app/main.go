package main

import (
	"OilStore/internal/cache"
	"OilStore/internal/config"
	"OilStore/internal/logger"
	"OilStore/internal/repository/postgres"
	"OilStore/internal/service"
	"OilStore/internal/transport"
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg := config.ConfigInit()
	logger.InitLogger(cfg.Loglvl, cfg.App_env)
	defer logger.CloseLogger()
	rdb := cache.RedisInit(cfg.Redis_addr)

	logger.Log.Info("✅ Redis готов")
	defer rdb.Close()
	ctx := context.Background()

	conn, errCon := postgres.ConnectionDBPostgres(ctx, cfg.Postgres_addr)

	if errCon != nil {
		logger.Log.Error("Ошибка подключения к БД", zap.Error(errCon))
		fmt.Println("Ошибка подключения к БД", errCon)
		return
	}
	logger.Log.Info("✅ PostgreSQL готов")
	defer conn.Close(ctx)

	oilRepo := postgres.NewOilConn(conn)
	oilServ := service.NewOilService(oilRepo, rdb)
	handlers := transport.NewHandlers(oilServ)

	rtr := transport.NewRouter(handlers)
	router := logger.MiddlewareLog(rtr)

	errBD := oilRepo.CreateTableOils(ctx)
	if errBD != nil {
		logger.Log.Error("❌ Ошибка при создании таблицы в БД", zap.Error(errBD))
		return
	}

	ctxStop, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	oilSrv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		logger.Log.Info("✅ Сервер запущен. Жду запросы на :8080")
		if errServer := oilSrv.ListenAndServe(); errServer != nil && errServer != http.ErrServerClosed {
			logger.Log.Error("Server Error!", zap.Error(errServer))
		}
	}()

	<-ctxStop.Done()
	logger.Log.Info("🛑 Получен сигнал остановки. Завершаю работу приложения и сервера...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := oilSrv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Info("❌ Ошибка остановки сервера...")
	}
	logger.Log.Info("⛔ Сервер завершил свою работу.")
}
