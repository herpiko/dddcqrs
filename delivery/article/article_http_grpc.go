package article_service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/herpiko/dddcqrs"
	event "github.com/herpiko/dddcqrs"
	uuid "github.com/satori/go.uuid"

	"github.com/herpiko/dddcqrs/domain/article"
	util "github.com/herpiko/dddcqrs/internal/util"

	"github.com/gorilla/mux"
)

func (ad *ArticleDelivery) HttpGrpcHandler(client event.EventStoreClient, router *mux.Router) {
	if client == nil {
		panic(errors.New("invalid-grpc-client"))
	}
	router.HandleFunc("/api/articles", ad.create(client)).Methods("POST")
	router.HandleFunc("/api/articles", ad.list(client)).Methods("GET")
}

func (ad *ArticleDelivery) create(client event.EventStoreClient) func(http.ResponseWriter, *http.Request) {
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

		jsonBytes, err := json.Marshal(articleItem)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusBadRequest, "invalid-payload")
			return
		}

		evID := uuid.NewV4()
		agID := uuid.NewV4()
		eventItem := &event.EventParam{
			Channel:       "article-created",
			EventType:     "article-created",
			AggregateType: "article",
			EventId:       evID.String(),
			AggregateId:   agID.String(),
			EventData:     string(jsonBytes),
		}
		x, _ := json.Marshal(eventItem)
		log.Println(string(x))
		res, err := client.CreateEvent(context.Background(), eventItem)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusInternalServerError, "internal-server-error")
			return
		}
		x, _ = json.Marshal(res)
		log.Println(string(x))
		util.Respond(w, http.StatusOK, nil)
		return
	}
	return handler
}

func (ad *ArticleDelivery) list(client event.EventStoreClient) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.FormValue("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.FormValue("limit"))
		if limit == 0 {
			limit = 10
		}
		search := r.FormValue("search")
		log.Println(search)
		param := &dddcqrs.Articles{
			Page:   int32(page),
			Limit:  int32(limit),
			Search: search,
		}
		_ = param

		jsonBytes, err := json.Marshal(param)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusBadRequest, "invalid-payload")
			return
		}

		evID := uuid.NewV4()
		agID := uuid.NewV4()
		eventItem := &event.EventParam{
			Channel:       "article-list",
			EventType:     "article-list",
			AggregateType: "article",
			EventId:       evID.String(),
			AggregateId:   agID.String(),
			EventData:     string(jsonBytes),
		}
		x, _ := json.Marshal(eventItem)
		log.Println(string(x))
		res, err := client.ListEvent(context.Background(), eventItem)
		log.Println(res)
		if err != nil {
			log.Println(err)
			util.RespondError(w, http.StatusInternalServerError, "internal-server-error")
			return
		}
		x, _ = json.Marshal(res)
		log.Println(string(x))
		util.RespondJson(w, http.StatusOK, res.EventData)
		return
	}
	return handler
}
