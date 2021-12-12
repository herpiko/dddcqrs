package main

import (
	"log"
	"os"
	"runtime"

	natsConn "github.com/herpiko/dddcqrs/conn/nats"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	// This is an event store where we store all events that happened in the mesh.
	// Should be hooked up to fast write database so that we can replay it again
	// in case something get wrong with a particular event
	// Didn't implement the database because it is out of the assesment scope / didn't have time
	sc := natsConn.Init()

	// Ensure that no other subscriber receive the same data,
	// let nats regulate the balancing.
	queueGroup := os.Getenv("SUBSCRIBER_QUEUE_GROUP")
	if queueGroup == "" {
		queueGroup = "default-queue-group"
	}
	sc.QueueSubscribe("event-store", os.Getenv("QUEUE_GROUP"), func(msg *nats.Msg) {
		// Simply print it
		log.Println(string(msg.Data))
	})

	log.Println("event-store running...")
	// Keep the connection alive
	runtime.Goexit()
}
