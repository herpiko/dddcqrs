package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"

	dddcqrs "github.com/herpiko/dddcqrs"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	delivery "github.com/herpiko/dddcqrs/delivery/article"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	elasticConn := el.NewElasticConn()
	articleDelivery, err := delivery.NewArticleDelivery(
		delivery.ArticleConfig(
			delivery.WithElastic(
				elasticConn.Conn,
			),
		),
	)
	_ = articleDelivery
	if err != nil {
		panic(err)
	}

	sc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	// Ensure that no other subscriber receive the same data
	// by using queue group. Let nats regulate the balancing.
	queueGroup := os.Getenv("SUBSCRIBER_QUEUE_GROUP")
	if queueGroup == "" {
		queueGroup = "default-queue-group"
	}
	sc.QueueSubscribe("article-list", os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		articleParam := &dddcqrs.Articles{}
		err := json.Unmarshal(msg.Data, &articleParam)
		if err != nil {
			log.Println(err)
		}
		x, _ := json.Marshal(articleParam)
		log.Println(string(x))

		articleList, err := articleDelivery.Articles.List(articleParam)
		if err != nil {
			log.Println(err)
		}

		log.Println(articleList)
		jsonBytes, err := json.Marshal(articleList)
		if err != nil {
			log.Println(err)
		}
		msg.Respond(jsonBytes)
	})

	log.Println("article-query running...")
	// Keep it alive
	runtime.Goexit()
}
