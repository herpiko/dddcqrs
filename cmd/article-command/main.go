package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	dddcqrs "github.com/herpiko/dddcqrs"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	natsConn "github.com/herpiko/dddcqrs/conn/nats"
	psql "github.com/herpiko/dddcqrs/conn/psql"
	delivery "github.com/herpiko/dddcqrs/delivery/article"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	articleDelivery, err := delivery.NewArticleDelivery(
		delivery.ArticleConfig(
			delivery.WithPsqlAndElastic(
				psql.NewPsqlConn().DB,
				el.NewElasticConn().Conn,
			),
		),
	)
	if err != nil {
		panic(err)
	}

	sc := natsConn.Init()

	// Ensure that no other subscriber receive the same data,
	// let nats regulate the balancing.
	queueGroup := os.Getenv("SUBSCRIBER_QUEUE_GROUP")
	if queueGroup == "" {
		queueGroup = "default-queue-group"
	}
	sc.QueueSubscribe("article-created", os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		log.Println("article-created")
		articleItem := &dddcqrs.Article{}
		err := json.Unmarshal(msg.Data, &articleItem)
		if err != nil {
			log.Println(err)
		}
		log.Println(articleItem)

		now := time.Now()
		articleItem.CreatedAt = now.Unix()

		// Write database: PSQL
		lastInsertId, err := articleDelivery.Articles.Create(articleItem)
		if err != nil {
			log.Println(err)
			msg.Respond([]byte(err.Error()))
			return
		}

		lastInsertIdStr := strconv.Itoa(int(lastInsertId))

		// Read database: Elastic
		err = articleDelivery.Articles.CreateAggregate(&dddcqrs.ArticleAggregateRoot{
			Id:        lastInsertIdStr,
			Title:     articleItem.Title,
			Body:      articleItem.Body,
			Author:    articleItem.Author,
			CreatedAt: now.Format(time.RFC3339),
		})
		if err != nil {
			log.Println(err)
			msg.Respond([]byte(err.Error()))
			return
		}

		// Any list cache should be obselete. Flush them
		sc.Publish("article-list-cache-flush", nil)
		msg.Respond([]byte(lastInsertIdStr))
	})

	log.Println("article-command running...")
	// Keep the connection alive
	runtime.Goexit()
}
