package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	event "github.com/herpiko/dddcqrs"
	article "github.com/herpiko/dddcqrs/delivery/article"
	"google.golang.org/grpc"

	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	router := mux.NewRouter()

	/*
		 * Article HTTP service
		 * Instead of initializing an article delivery instance with direct access to db,
		 * we want to talk to a GRPC service instead
			app := dddcqrs.NewApp()
			as, err := article.NewArticleDelivery(
				article.ArticleConfig(
					article.ArticleRepoWithPsql(app.DB),
				),
			)
			if err != nil {
				panic(err)
			}
			as.HttpHandler(router)
	*/

	// Article HTTP service with GRPC backend
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := event.NewEventStoreClient(conn)
	//list := event.NewAddServiceClient(conn)
	as, err := article.NewArticleDelivery()
	if err != nil {
		panic(err)
	}
	as.HttpGrpcHandler(client, router)
	if err != nil {
		panic(err)
	}

	addr := "0.0.0.0:8000"
	log.Println("http running at " + addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
