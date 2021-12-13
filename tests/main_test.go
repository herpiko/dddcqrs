package test

import (
	"context"
	"log"
	"os"
	"testing"

	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	natsConn "github.com/herpiko/dddcqrs/conn/nats"
	psql "github.com/herpiko/dddcqrs/conn/psql"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func TestMain(m *testing.M) {
	os.Setenv("MIGRATION_PATH", "../")
	os.Setenv("TEST_USING_CACHED_DATA", "false")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	db := psql.NewPsqlConn().DB
	defer db.Close()

	elastic := el.NewElasticConn().Conn
	_, _ = elastic.DeleteIndex("article").Do(context.Background())

	time.Sleep(1000 * time.Millisecond)

	_, err := db.Exec("DELETE FROM articles")
	if err != nil {
		panic(err)
	}

	sc := natsConn.Init()
	sc.Subscribe("test", func(msg *nats.Msg) {
		if string(msg.Data) == "use-cached-data" {
			os.Setenv("TEST_USING_CACHED_DATA", "true")
		} else {
			os.Setenv("TEST_USING_CACHED_DATA", "false")
		}
	})

	defer sc.Close()

	code := m.Run()
	os.Exit(code)
}
