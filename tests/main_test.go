package test

import (
	"log"
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	psql "github.com/herpiko/dddcqrs/conn/psql"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	os.Setenv("MIGRATION_PATH", "../")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	db := psql.NewPsqlConn().DB
	defer db.Close()

	elastic := el.NewElasticConn().Conn
	elastic.Flush()

	_, err := db.Exec("DELETE FROM articles")
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}
