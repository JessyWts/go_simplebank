DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

network:
	docker network create bank-network

start_postgres:
	docker run --name postgres16 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

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

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go bitbucket.org/jessyw/go_simplebank/db/sqlc Store

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

# test gRPC server
evans:
	evans --host localhost --port 9090 -r repl

.PHONY: network start_postgres stop_postgres delete_postgres create_db drop_db migrate_up migrate_down migrate_up_last migrate_down_last new_migration db_docs db_schema sqlc test server mock proto evans