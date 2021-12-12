package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	dddcqrs "github.com/herpiko/dddcqrs"
	article "github.com/herpiko/dddcqrs/delivery/article"
	"google.golang.org/grpc"

	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	router := mux.NewRouter()

	grpcAddr := os.Getenv("GRPC_ADDRESS")
	if grpcAddr == "" {
		grpcAddr = "localhost:4040"
	}

	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := dddcqrs.NewArticleServiceClient(conn)

	// Init without storage, we only need its HTTP handler
	as, err := article.NewArticleDelivery()
	if err != nil {
		panic(err)
	}

	// Article HTTP handler with GRPC backend
	as.HttpGrpcHandler(client, router)

	addr := "0.0.0.0:8000"
	log.Println("http running at " + addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
