package config

import "os"

type Config struct {
	Postgres_addr string
	Redis_addr    string
	Loglvl        string
	App_env       string
}

func ConfigInit() *Config {
	return &Config{
		Postgres_addr: os.Getenv("CONSTR"),
		Redis_addr:    os.Getenv("REDIS_ADDR"),
		Loglvl:        os.Getenv("LOGLVL"),
		App_env:       os.Getenv("APP_ENV"),
	}
}
