package article_service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	dddcqrs "github.com/herpiko/dddcqrs"

	"github.com/herpiko/dddcqrs/domain/article"
	util "github.com/herpiko/dddcqrs/internal/util"

	"github.com/gorilla/mux"
)

// HTTP handler that talk to GRPC backend

func (ad *ArticleDelivery) HttpGrpcHandler(client dddcqrs.ArticleServiceClient, router *mux.Router) {
	/*
		if client.Event == nil {
			panic(errors.New("invalid-grpc-client"))
		}
		if client == nil {
			panic(errors.New("invalid-grpc-client"))
		}
	*/
	router.HandleFunc("/api/articles", ad.create(client)).Methods("POST")
	router.HandleFunc("/api/articles", ad.list(client)).Methods("GET")
	router.HandleFunc("/api/article/{id}", ad.get(client)).Methods("GET")
}

func (ad *ArticleDelivery) create(client dddcqrs.ArticleServiceClient) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var item dddcqrs.Article
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&item); err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusBadRequest, "invalid-payload")
			return
		}
		defer r.Body.Close()

		// This help to validate
		articleItem, err := article.NewArticle(item.Title, item.Body, item.Author)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusBadRequest, "invalid-payload")
			return
		}

		res, err := client.CreateArticle(context.Background(), articleItem)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusInternalServerError, "internal-server-error")
			return
		}
		util.Respond(w, http.StatusOK, res)
		return

	}
	return handler
}

func (ad *ArticleDelivery) list(client dddcqrs.ArticleServiceClient) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.FormValue("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.FormValue("limit"))
		if limit == 0 {
			limit = 10
		}
		search := r.FormValue("query")  // title and body
		author := r.FormValue("author") // title and body

		param := &dddcqrs.Articles{
			Page:          int32(page),
			Limit:         int32(limit),
			ArticleFilter: search,
			AuthorFilter:  author,
		}
		_ = param
		res, err := client.ListArticle(context.Background(), param)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusInternalServerError, "internal-server-error")
			return
		}
		util.RespondJson(w, http.StatusOK, res.Data)
		return
	}
	return handler
}

func (ad *ArticleDelivery) get(client dddcqrs.ArticleServiceClient) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		res, err := client.GetArticle(context.Background(), &dddcqrs.ArticleId{Id: id})
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusInternalServerError, "internal-server-error")
			return
		}
		if strings.Contains(res.Data, "not-found") {
			util.RespondError(w, http.StatusNotFound, "not-found")
			return
		}
		util.RespondJson(w, http.StatusOK, res.Data)
		return
	}
	return handler
}
