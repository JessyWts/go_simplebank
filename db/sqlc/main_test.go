package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *pgx.Conn

// func TestMain(m *testing.M) {
// 	conn, err := pgx.Connect(context.Background(), dbSource)
// 	if err != nil {
// 		log.Fatal("cannot connect to db:", err)
// 	}
// 	testQueries = New(conn)

// 	os.Exit(m.Run())
// }

func TestMain(m *testing.M) {
	var err error
	testDB, err = pgx.Connect(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	defer testDB.Close(context.Background())

	testQueries = New(testDB)

	os.Exit(m.Run())
}
