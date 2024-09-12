DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

start_postgres:
	docker run --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

stop_postgres:
	docker stop postgres16

delete_postgres:
	docker stop postgres16 && docker rm postgres16

create_db:
	docker exec -it postgres16 createdb --username=root --owner=root simple_bank

drop_db:
	docker exec -it postgres16 dropdb simple_bank

migrate_up:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrate_up_last:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migrate_down:
	migrate -path db/migration -database "$(DB_URL)" -verbose down -all

migrate_down_last:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go bitbucket.org/jessyw/go_simplebank/db/sqlc Store

.PHONY: start_postgres stop_postgres delete_postgres create_db drop_db migrate_up migrate_down migrate_up_last migrate_down_last sqlc test server mock