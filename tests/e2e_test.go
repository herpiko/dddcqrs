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

	os.Setenv("TEST_USING_CACHED_DATA", "false")
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
	assert.Equal(t, "false", os.Getenv("TEST_USING_CACHED_DATA"))

	// Get, cached
	req, _ = http.NewRequest("GET", "/api/article/"+articleId.Id, nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleItem = &dddcqrs.ArticleAggregateRoot{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleItem)
	assert.Equal(t, completeArticleItem.Title, articleItem.Title)
	assert.Equal(t, completeArticleItem.Body, articleItem.Body)
	assert.Equal(t, completeArticleItem.Author, articleItem.Author)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "true", os.Getenv("TEST_USING_CACHED_DATA"))
	os.Setenv("TEST_USING_CACHED_DATA", "false")

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

	type testCase struct {
		test        string
		article     *dddcqrs.Article
		expectedErr error
	}

	testCases := []testCase{
		{
			test: "Article 1",
			article: &dddcqrs.Article{
				Title:  "robohnya surau kami",
				Body:   "tentang orang yang lupa hablumminannas",
				Author: "aa navis",
			},
			expectedErr: nil,
		},
		{
			test: "Article 2",
			article: &dddcqrs.Article{
				Title:  "radical candor",
				Body:   "being honest is good",
				Author: "kim scott",
			},
			expectedErr: nil,
		},
		{
			test: "Article 3",
			article: &dddcqrs.Article{
				Title:  "metamorfosis",
				Body:   "the longest short story from kafka",
				Author: "franz kafka",
			},
			expectedErr: nil,
		},
		{
			test: "Article 4",
			article: &dddcqrs.Article{
				Title:  "the rosie project",
				Body:   "back at the bar",
				Author: "graeme simsion",
			},
			expectedErr: nil,
		},
		{
			test: "Article 5",
			article: &dddcqrs.Article{
				Title:  "moby dick",
				Body:   "epic saga of one legend fanatic",
				Author: "herman melville",
			},
			expectedErr: nil,
		},
		{
			test: "Article 6",
			article: &dddcqrs.Article{
				Title:  "the name of the rose",
				Body:   "imagine a medieval castle",
				Author: "umberto uco",
			},
			expectedErr: nil,
		},
		{
			test: "Article 7",
			article: &dddcqrs.Article{
				Title:  "scandal",
				Body:   "when suguro and kurimoto arrived",
				Author: "endo",
			},
			expectedErr: nil,
		},
		{
			test: "Article 8",
			article: &dddcqrs.Article{
				Title:  "the alchemist",
				Body:   "the boys name was santiago",
				Author: "paulo coelho",
			},
			expectedErr: nil,
		},
		{
			test: "Article 9",
			article: &dddcqrs.Article{
				Title:  "momo",
				Body:   "kisah momo berlangsung di negeri khayalan",
				Author: "michael ende",
			},
			expectedErr: nil,
		},
		{
			test: "Article 10",
			article: &dddcqrs.Article{
				Title:  "3 years",
				Body:   "jatuh cinta kepada yulia",
				Author: "anton chekov",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		// Need to wait a bit so we'll have chronological sequence
		time.Sleep(1000 * time.Millisecond)
		t.Log(tc.article.Title)
		t.Run(tc.test, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tc.article)
			req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			assert.Equal(t, http.StatusOK, response.Code)
		})
	}

	// Wait a bit so we can determine that this one is the newest article
	time.Sleep(1100 * time.Millisecond)
	newestArticleItem := dddcqrs.ArticleAggregateRoot{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	jsonBytes, _ := json.Marshal(&newestArticleItem)
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

	time.Sleep(1000 * time.Millisecond)

	// Get list
	req, _ = http.NewRequest("GET", "/api/articles", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList := &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 10, len(articleList.Data))
	assert.Equal(t, newestArticleItem.Title, articleList.Data[0].Title)

	// Get page 2 / limit 2
	req, _ = http.NewRequest("GET", "/api/articles?page=2&limit=2", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 2, len(articleList.Data))
	assert.Equal(t, "momo", articleList.Data[0].Title)
	assert.Equal(t, "the alchemist", articleList.Data[1].Title)

	// Search by title
	req, _ = http.NewRequest("GET", "/api/articles?query=surau", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "robohnya surau kami", articleList.Data[0].Title)
	assert.Equal(t, "false", os.Getenv("TEST_USING_CACHED_DATA"))

	// Search by title, cached
	req, _ = http.NewRequest("GET", "/api/articles?query=surau", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "robohnya surau kami", articleList.Data[0].Title)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "true", os.Getenv("TEST_USING_CACHED_DATA"))
	os.Setenv("TEST_USING_CACHED_DATA", "false")

	// Search by body
	req, _ = http.NewRequest("GET", "/api/articles?query=suguro", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "scandal", articleList.Data[0].Title)
	assert.Equal(t, "false", os.Getenv("TEST_USING_CACHED_DATA"))

	// Search by body, cached
	req, _ = http.NewRequest("GET", "/api/articles?query=suguro", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "scandal", articleList.Data[0].Title)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "true", os.Getenv("TEST_USING_CACHED_DATA"))
	os.Setenv("TEST_USING_CACHED_DATA", "false")

	// Search by author
	req, _ = http.NewRequest("GET", "/api/articles?author=umberto", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "the name of the rose", articleList.Data[0].Title)
	assert.Equal(t, "false", os.Getenv("TEST_USING_CACHED_DATA"))

	// Search by author, cached
	req, _ = http.NewRequest("GET", "/api/articles?author=umberto", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "the name of the rose", articleList.Data[0].Title)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "true", os.Getenv("TEST_USING_CACHED_DATA"))

	// Create another article, list cache will be flushed
	os.Setenv("TEST_USING_CACHED_DATA", "false")
	completeArticleItem := dddcqrs.ArticleAggregateRoot{
		Title:  "ok",
		Body:   "ok",
		Author: "ok",
	}
	jsonBytes, _ = json.Marshal(&completeArticleItem)
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	time.Sleep(1000 * time.Millisecond)

	// Search by author, fresh
	req, _ = http.NewRequest("GET", "/api/articles?author=umberto", nil)
	req.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)
	articleList = &dddcqrs.Articles{}
	jsonBytes = response.Body.Bytes()
	json.Unmarshal(jsonBytes, articleList)
	assert.Equal(t, 1, len(articleList.Data))
	assert.Equal(t, "the name of the rose", articleList.Data[0].Title)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "false", os.Getenv("TEST_USING_CACHED_DATA"))
}
