package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
)

func RespondError(w http.ResponseWriter, code int, message string) {
	Respond(w, code, map[string]string{"error": message})
}

func Respond(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func MigrateInit(db *sql.DB) error {
	_, b, _, _ := runtime.Caller(0)
	cwd := filepath.Dir(b)
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
	}
	err = migration.Drop()
	if err != nil && err.Error() != "no change" {
		log.Println(err)
	}
	_, _ = migration.Close()
	migration, err = migrate.New(
		migrationPath,
		connectionString)
	if err != nil {
		log.Println(err)
	}
	err = migration.Up()
	if err != nil && err.Error() != "no change" {
		log.Println(err)
	}
	_, _ = migration.Close()

	// Backup migrated db
	connectionString =
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			"5432",
			"postgres",
		)
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	_, _ = db.Exec(`
	DROP DATABASE ` + os.Getenv("DB_NAME") + `_ready`)
	_, err = db.Exec(`
	CREATE DATABASE ` + os.Getenv("DB_NAME") + `_ready WITH TEMPLATE ` + os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatal(err)
		return err
	}

	db.Close()
	return nil
}

func MigrateClean(db *sql.DB) error {
	db.Close()
	connectionString :=
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			"postgres",
		)
	var err error

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = db.Exec("DROP DATABASE " + os.Getenv("DB_NAME"))
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = db.Exec("CREATE DATABASE " + os.Getenv("DB_NAME") + " WITH TEMPLATE " + os.Getenv("DB_NAME") + "_ready")
	if err != nil {
		log.Println(err)
		return err
	}

	// Reconnect again using testpsql
	db.Close()
	connectionString =
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			"testpsql",
		)

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
