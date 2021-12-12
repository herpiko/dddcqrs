package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	dddcqrs "github.com/herpiko/dddcqrs"
	article "github.com/herpiko/dddcqrs/delivery/article"
	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
)

func TestCreateArticle(t *testing.T) {
	router := mux.NewRouter()
	grpcAddr := os.Getenv("GRPC_ADDRESS")
	if grpcAddr == "" {
		grpcAddr = "localhost:4040"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	assert.Equal(t, err, nil)
	client := dddcqrs.NewArticleServiceClient(conn)
	as, err := article.NewArticleDelivery()
	assert.Equal(t, err, nil)
	as.HttpGrpcHandler(client, router)
	assert.Equal(t, err, nil)

	// Invalid JSON
	jsonBytes := []byte(`{"title":pygmy","body":"nicknamed pygmy for his diminutive size","author":"chuck palahniuk"}`)
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Empty JSON
	jsonBytes = []byte(`{}`)
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Empty title
	jsonBytes = []byte(`{"body":"nicknamed pygmy for his diminutive size","author":"chuck palahniuk"}`)
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Empty body
	jsonBytes = []byte(`{"title":"pygmy","author":"chuck palahniuk"}`)
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Empty author
	jsonBytes = []byte(`{"title":"pygmy","body":"nicknamed pygmy for his diminutive size"}`)
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Complete
	completeArticleItem := dddcqrs.ArticleAggregateRoot{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	jsonBytes, _ = json.Marshal(&completeArticleItem)
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	var articleId dddcqrs.ArticleId
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, &articleId)

	time.Sleep(1000 * time.Millisecond)

	// Confirm it
	req, _ = http.NewRequest("GET", "/api/article/"+articleId.Id, nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleItem := &dddcqrs.ArticleAggregateRoot{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleItem)
	assert.Equal(t, completeArticleItem.Title, articleItem.Title)
	assert.Equal(t, completeArticleItem.Body, articleItem.Body)
	assert.Equal(t, completeArticleItem.Author, articleItem.Author)
}

func TestGetArticle(t *testing.T) {
	router := mux.NewRouter()
	grpcAddr := os.Getenv("GRPC_ADDRESS")
	if grpcAddr == "" {
		grpcAddr = "localhost:4040"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	assert.Equal(t, err, nil)
	client := dddcqrs.NewArticleServiceClient(conn)
	as, err := article.NewArticleDelivery()
	assert.Equal(t, err, nil)
	as.HttpGrpcHandler(client, router)
	assert.Equal(t, err, nil)

	// Create
	completeArticleItem := dddcqrs.ArticleAggregateRoot{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	jsonBytes, _ := json.Marshal(&completeArticleItem)
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	var articleId dddcqrs.ArticleId
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, &articleId)

	time.Sleep(1000 * time.Millisecond)

	// Get
	req, _ = http.NewRequest("GET", "/api/article/"+articleId.Id, nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleItem := &dddcqrs.ArticleAggregateRoot{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleItem)
	assert.Equal(t, completeArticleItem.Title, articleItem.Title)
	assert.Equal(t, completeArticleItem.Body, articleItem.Body)
	assert.Equal(t, completeArticleItem.Author, articleItem.Author)

	// Invalid id
	req, _ = http.NewRequest("GET", "/api/article/123456789", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestListArticle(t *testing.T) {
	router := mux.NewRouter()
	grpcAddr := os.Getenv("GRPC_ADDRESS")
	if grpcAddr == "" {
		grpcAddr = "localhost:4040"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	assert.Equal(t, err, nil)
	client := dddcqrs.NewArticleServiceClient(conn)
	as, err := article.NewArticleDelivery()
	assert.Equal(t, err, nil)
	as.HttpGrpcHandler(client, router)
	assert.Equal(t, err, nil)

	// Create
	completeArticleItem := dddcqrs.ArticleAggregateRoot{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	jsonBytes, _ := json.Marshal(&completeArticleItem)
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	var articleId dddcqrs.ArticleId
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, &articleId)

	time.Sleep(1000 * time.Millisecond)

	// Get list
	req, _ = http.NewRequest("GET", "/api/articles", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
}
