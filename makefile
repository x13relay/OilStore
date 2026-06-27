include .env
export
export PROJECT_ROOT=$(shell pwd)
env_up:
	docker compose up -d
env_cleanup:
	@read -p "Очистить все volumе файлы окружения? [y/N]: " ans;\
	if ["$$ans"="y" ]; then \
	docker compose down && \
	rm -rf out/pgdata && \
	echo "файлы и папки окружения удалены"; \
	else \
	echo "очистка окружения отменена"; \
	fi

test_run:
	go run cmd/app/main.go
test_run_race:
	go run -race cmd/app/main.go

db_start:
	docker run -d --name postgres-dev \
	  -e POSTGRES_USER=${POSTGRES_USER} \
	  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	  -e POSTGRES_DB=${POSTGRES_DB} \
	  -p 5432:5432 \
	  -v ${PROJECT_ROOT}/out/pgdata:/var/lib/postgresql/data \
	  postgres:17-alpine
db_stop:
	docker stop postgres-dev || true
	docker rm postgres-dev || true
redis_start:
	docker run -d --name redis-dev -p 6379:6379 redis:alpine
redis_stop:
	docker stop redis-dev || true
	docker rm redis-dev || true
migrate_create:
	migrate create -ext sql -dir migrations -seq init_oilstore
migrate_up:
	migrate -path migrations -database "$(CONSTR)" up
migrate_down:
	migrate -path migrations -database "$(CONSTR)" down 1
.PHONY: env_up test_run test_run_race db_start db_stop redis_start redis_stop migrate_create migrate_up migrate_down

