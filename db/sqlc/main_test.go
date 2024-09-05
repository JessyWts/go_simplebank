package db

import (
	"context"
	"log"
	"os"
	"testing"

	"bitbucket.org/jessyw/go_simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	defer connPool.Close()

	testStore = NewStore(connPool)

	os.Exit(m.Run())
}
