package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log     *zap.Logger
	logFile *os.File
)

func CloseLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
	if logFile != nil {
		_ = logFile.Close()
	}
}
func InitLogger(env, levelStr string) {
	//env := os.Getenv("APP_ENV")
	//levelStr := os.Getenv("LOGLVL")

	logFile, errFile := os.OpenFile("cmd/app/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errFile != nil {
		panic(errFile)
	}

	consoleSyncer := zapcore.AddSync(os.Stdout)
	fileSyncer := zapcore.AddSync(logFile)

	var encoder zapcore.Encoder
	var consoleEncoder zapcore.Encoder

	//var cfg zap.Config

	if env == "production" {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		consoleEncoder = encoder
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
		devCfg := zap.NewDevelopmentEncoderConfig()
		devCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(devCfg)
	}

	var lvl zapcore.Level

	if err := lvl.UnmarshalText([]byte(levelStr)); err != nil {
		lvl = zapcore.InfoLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleSyncer, lvl),
		zapcore.NewCore(encoder, fileSyncer, lvl))

	Log = zap.New(core)

	//cfg.Level = zap.NewAtomicLevelAt(lvl)
	//Log, _ = cfg.Build()

}
