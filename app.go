package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Init(host, user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	log.Println(connectionString)
	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initRoutes()
}

func (a *App) MigrateInit() error {
	cwd, _ := os.Getwd()
	migrationPath := "file://" + cwd + "/migrations"
	connectionString :=
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			"localhost",
			"5432",
			os.Getenv("DB_NAME"),
		)

	log.Println(fmt.Sprintf("Migrating %s from %s", connectionString, migrationPath))
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
			"localhost",
			"5432",
			"postgres",
		)
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	_, _ = a.DB.Exec(`
	DROP DATABASE ` + os.Getenv("DB_NAME") + `_ready`)
	_, err = a.DB.Exec(`
	CREATE DATABASE ` + os.Getenv("DB_NAME") + `_ready WITH TEMPLATE ` + os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatal(err)
		return err
	}

	a.DB.Close()
	return nil
}

func (a *App) MigrateClean() error {
	a.DB.Close()
	connectionString :=
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			"postgres",
		)
	var err error

	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = a.DB.Exec("DROP DATABASE " + os.Getenv("DB_NAME"))
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = a.DB.Exec("CREATE DATABASE " + os.Getenv("DB_NAME") + " WITH TEMPLATE " + os.Getenv("DB_NAME") + "_ready")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (a *App) Run(addr string) {
	log.Println("Running on port ", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (a *App) initRoutes() {
	a.Router.HandleFunc("/", a.indexHandler).Methods("GET")
	a.Router.HandleFunc("/api/projects", a.getProjects).Methods("GET")
	a.Router.HandleFunc("/api/project", a.createProject).Methods("POST")
	a.Router.HandleFunc("/api/project/{id}", a.getProject).Methods("GET")
	a.Router.HandleFunc("/api/project/{id}", a.updateProject).Methods("PUT")
	a.Router.HandleFunc("/api/project/{id}", a.deleteProject).Methods("DELETE")
}
