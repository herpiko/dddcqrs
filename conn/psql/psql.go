package dddcqrs

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Singleton, will be used by command instance
type PsqlConn struct {
	DB *sql.DB
}

var (
	psqlConn *PsqlConn
)
var lock = &sync.Mutex{}

func NewPsqlConn() *PsqlConn {
	lock.Lock()
	defer lock.Unlock()
	if psqlConn == nil {
		psqlConn = &PsqlConn{}
	}
	// Make sure our DB has the latest migrations
	psqlConn.init()
	return psqlConn
}

func (psqlConn *PsqlConn) init() {
	var err error

	// Regular migration
	err = psqlConn.migrateUp()
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	// Main database connection
	connectionString :=
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_NAME"),
		)
	psqlConn.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
}

func (psqlConn *PsqlConn) migrateUp() error {
	cwd, _ := os.Getwd()
	migrationPath := "file://" + cwd + "/migrations"
	connectionString :=
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			"5432",
			os.Getenv("DB_NAME"),
		)

	var err error

	migration, err := migrate.New(
		migrationPath,
		connectionString)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(os.Getenv("MIGRATE_FORCE")) > 0 {
		// Force specific migration version
		migrateVersion, err := strconv.Atoi(os.Getenv("MIGRATE_FORCE"))
		if err != nil {
			log.Println(err)
			return err
		}
		err = migration.Force(migrateVersion)
		if err != nil && err.Error() != "no change" {
			log.Println(err)
			return err
		}
	} else {
		// Regular migration
		err = migration.Up()
		if err != nil && err.Error() != "no change" {
			log.Println(err)
			return err
		}
	}

	_, err = migration.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
