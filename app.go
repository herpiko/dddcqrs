package dddcqrs

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Singleton, will be used by command instance
type App struct {
	DB *sql.DB
}

var (
	app *App
)
var lock = &sync.Mutex{}

func NewApp() *App {
	lock.Lock()
	defer lock.Unlock()
	if app == nil {
		app = &App{}
	}
	// Make sure our DB has the latest migrations
	app.init()
	return app
}

func (app *App) init() {
	var err error

	// Regular migration
	err = app.migrateUp()
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
	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
}

func (app *App) migrateUp() error {
	_, b, _, _ := runtime.Caller(0)
	cwd := filepath.Dir(b)
	migrationPath := "file://" + cwd + "/migrations"
	log.Println(migrationPath)
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
