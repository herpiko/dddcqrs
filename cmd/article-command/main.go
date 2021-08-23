package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"runtime"

	dddcqrs "github.com/herpiko/dddcqrs"
	"github.com/herpiko/dddcqrs/domain/article/psql"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

const (
	subscribeChannel = "article-created"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	app := dddcqrs.NewApp()
	articleRepo, err := psql.New(context.Background(), app.DB)
	if err != nil {
		panic(err)
	}

	sc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	// Ensure that no other subscriber receive the same data,
	// let nats regulate the balancing.
	queueGroup := os.Getenv("SUBSCRIBER_QUEUE_GROUP")
	if queueGroup == "" {
		queueGroup = "default-queue-group"
	}
	sc.QueueSubscribe(subscribeChannel, os.Getenv("QUEUE_GROUP"), func(m *nats.Msg) {
		article := &dddcqrs.Article{}
		err := json.Unmarshal(m.Data, &article)
		if err != nil {
			log.Println(err)
		}
		x, _ := json.Marshal(article)
		log.Println(string(x))

		// TODO create/store to db
		err = articleRepo.Create(article)
		if err != nil {
			log.Println(err)
		}
	})

	log.Println("article-command running...")
	// Keep the connection alive
	runtime.Goexit()
}
