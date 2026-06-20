include .env
export

test_run:
	go run cmd/app/main.go
test_run_race:
	go run -race main.go
oil_start:
	docker compose up -d 
oil_stop:
	docker compose down 
all_stop:
	docker copmose down
db_start:
	docker run -d --name postgres-dev \
	  -e POSTGRES_USER=${POSTGRES_USER} \
	  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	  -e POSTGRES_DB=${POSTGRES_DB} \
	  -p 5432:5432 \
	  -v httpgohandlers-postgresql_postgres_data:/var/lib/postgresql/data \
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

