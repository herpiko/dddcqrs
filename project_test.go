package main

import (
	"log"
	"os"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/joho/godotenv"
)

var a App

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()
	a = App{}
	a.MigrateInit()
	a.Init(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"))
	code := m.Run()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	a.MigrateClean()

	req, _ := http.NewRequest("GET", "/api/projects", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProject(t *testing.T) {
	a.MigrateClean()

	req, _ := http.NewRequest("GET", "/api/project/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Project not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Project not found'. Got '%s'", m["error"])
	}
}

// tom: rewritten function
func TestCreateProject(t *testing.T) {

	a.MigrateClean()

	var jsonStr = []byte(`{"name":"test project"}`)
	req, _ := http.NewRequest("POST", "/api/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test project" {
		t.Errorf("Expected project name to be 'test project'. Got '%v'", m["name"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected project ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProject(t *testing.T) {
	a.MigrateClean()
	addProjects(1)

	req, _ := http.NewRequest("GET", "/api/project/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProjects(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO projects(name) VALUES($1)", "Project "+strconv.Itoa(i))
	}
}

func TestUpdateProject(t *testing.T) {

	a.MigrateClean()
	addProjects(1)

	req, _ := http.NewRequest("GET", "/api/project/1", nil)
	response := executeRequest(req)
	var originalProject map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProject)

	var jsonStr = []byte(`{"name":"test project - updated name"}`)
	req, _ = http.NewRequest("PUT", "/api/project/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	// req, _ = http.NewRequest("PUT", "/api/project/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProject["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProject["id"], m["id"])
	}

	if m["name"] == originalProject["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProject["name"], m["name"], m["name"])
	}
}

func TestDeleteProject(t *testing.T) {
	a.MigrateClean()
	addProjects(1)

	req, _ := http.NewRequest("GET", "/api/project/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/project/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/project/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
