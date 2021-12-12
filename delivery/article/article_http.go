package article_service

import (
	"errors"
	"net/http"

	util "github.com/herpiko/dddcqrs/internal/util"

	"github.com/gorilla/mux"
)

// An example of HTTP handler with direct access to repo/db without grpc/mesh
func (ad *ArticleDelivery) HttpHandler(router *mux.Router) {
	if ad.Articles == nil {
		panic(errors.New("uninitialized-repo"))
	}
	router.HandleFunc("/api/articles", ad.getAll).Methods("GET")
}

func (ad *ArticleDelivery) getAll(w http.ResponseWriter, r *http.Request) {
	util.Respond(w, http.StatusOK, nil)
}
