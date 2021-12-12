package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"

	dddcqrs "github.com/herpiko/dddcqrs"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	psql "github.com/herpiko/dddcqrs/conn/psql"
	delivery "github.com/herpiko/dddcqrs/delivery/article"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

const (
	subscribeChannel = "article-created"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	elasticConn := el.NewElasticConn()
	psqlConn := psql.NewPsqlConn()
	articleDelivery, err := delivery.NewArticleDelivery(
		delivery.ArticleConfig(
			delivery.WithPsqlAndElastic(
				psqlConn.DB,
				elasticConn.Conn,
			),
		),
	)
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
	sc.QueueSubscribe(subscribeChannel, os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		articleItem := &dddcqrs.Article{}
		err := json.Unmarshal(msg.Data, &articleItem)
		if err != nil {
			log.Println(err)
		}
		x, _ := json.Marshal(articleItem)
		log.Println(string(x))

		// Write database
		err = articleDelivery.Articles.Create(articleItem)
		if err != nil {
			log.Println(err)
			msg.Respond([]byte(err.Error()))
			return
		}

		// Read database
		err = articleDelivery.Articles.CreateAggregate(&dddcqrs.ArticleAggregateRoot{
			Title:      articleItem.Title,
			Body:       articleItem.Body,
			AuthorName: articleItem.Author,
		})
		if err != nil {
			log.Println(err)
			msg.Respond([]byte(err.Error()))
			return
		}
		msg.Respond(nil)
	})

	log.Println("article-command running...")
	// Keep the connection alive
	runtime.Goexit()
}
