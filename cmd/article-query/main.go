package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/allegro/bigcache/v3"
	dddcqrs "github.com/herpiko/dddcqrs"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	natsConn "github.com/herpiko/dddcqrs/conn/nats"
	delivery "github.com/herpiko/dddcqrs/delivery/article"
	util "github.com/herpiko/dddcqrs/internal/util"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	articleDelivery, err := delivery.NewArticleDelivery(
		delivery.ArticleConfig(
			delivery.WithElastic(
				el.NewElasticConn().Conn,
			),
		),
	)
	_ = articleDelivery
	if err != nil {
		panic(err)
	}

	sc := natsConn.Init()

	// Elasticsearch is already providing cache but it still has
	// cost on network. Use bigcache:
	// - for it's performance
	// - as fallback in case something get wrong with elastic

	// Separated list and item cache since we want to
	// flush the entire list cache whenever new item get added
	listCache, _ := bigcache.NewBigCache(
		bigcache.DefaultConfig(168 * time.Hour), // A week
	)
	itemCache, _ := bigcache.NewBigCache(
		bigcache.DefaultConfig(168 * time.Hour), // A week
	)

	// Ensure that no other subscriber receive the same data
	// by using queue group. Let nats regulate the balancing.
	queueGroup := os.Getenv("SUBSCRIBER_QUEUE_GROUP")
	if queueGroup == "" {
		queueGroup = "default-queue-group"
	}

	// Event listener

	sc.QueueSubscribe("article-list", os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		log.Println("article-list")
		articleParam := &dddcqrs.Articles{}
		err := json.Unmarshal(msg.Data, &articleParam)
		if err != nil {
			log.Println(err)
		}

		// Use hashed string of param config as key.
		// Marshaled struct/proto is always consistent,
		// we can trust it.
		cacheId := util.Sha256sum(string(msg.Data))
		cached, _ := listCache.Get(cacheId)
		if len(cached) > 0 {
			sc.Publish("test", []byte("use-cached-data"))
			log.Println("Use cached data")
			msg.Respond(cached)
			return
		}
		sc.Publish("test", []byte("use-fresh-data"))
		log.Println("Use fresh data")

		articleList, err := articleDelivery.Articles.List(articleParam)
		if err != nil {
			log.Println(err)
		}

		log.Println(articleList)
		jsonBytes, err := json.Marshal(articleList)
		if err != nil {
			log.Println(err)
		}
		listCache.Set(cacheId, jsonBytes)
		msg.Respond(jsonBytes)
	})

	sc.QueueSubscribe("article-get", os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		log.Println("article-get")
		id := string(msg.Data)
		cacheId := "article-get-" + id
		cached, _ := itemCache.Get(cacheId)
		if len(cached) > 0 {
			sc.Publish("test", []byte("use-cached-data"))
			log.Println("Use cached data")
			msg.Respond(cached)
			return
		}
		sc.Publish("test", []byte("use-fresh-data"))
		log.Println("Use fresh data")

		articleItem, err := articleDelivery.Articles.Get(id)
		if err != nil {
			log.Println(err)
			msg.Respond([]byte(err.Error()))
			return
		}

		log.Println(articleItem)
		jsonBytes, err := json.Marshal(articleItem)
		if err != nil {
			log.Println(err)
			msg.Respond([]byte(err.Error()))
			return
		}

		itemCache.Set(cacheId, jsonBytes)
		msg.Respond(jsonBytes)
	})

	// Flush list cache whenever new article get added
	sc.QueueSubscribe("article-list-cache-flush", os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		listCache.Reset()
		log.Println("Article list cache flushed")
	})

	log.Println("article-query running...")
	// Keep it alive
	runtime.Goexit()
}
