DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

start_postgres:
	docker run --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

stop_postgres:
	docker stop postgres16

delete_postgres:
	docker stop postgres16 && docker rm postgres16

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres16 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

.PHONY: start_postgres stop_postgres delete_postgres createdb dropdb migrateup migratedown sqlc test server