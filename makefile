test_run:
	go run main.go
oilstore_start:
	docker compose up -d oilstore
oilstore_stop:
	docker compose down oilstore
all_stop:
	docker copmose down
bd_start:
	docker run -d --name postgres-dev \
	  -e POSTGRES_USER=${POSTGRES_USER} \
	  -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	  -e POSTGRES_DB=${POSTGRES_DB} \
	  -p 5432:5432 \
	  -v httpgohandlers-postgresql_postgres_data:/var/lib/postgresql/data \
	  postgres:17-alpine
bd_stop:
	docker stop postgres-dev || true
	docker rm postgres-dev || true