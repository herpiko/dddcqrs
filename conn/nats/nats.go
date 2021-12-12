package nats

import (
	"os"
	"time"

	event "github.com/herpiko/dddcqrs"
	"github.com/nats-io/nats.go"
)

type Stream struct{}

var sc *nats.Conn

func Init() *nats.Conn { // init helper
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://127.0.0.1:4222"
	}
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	return conn
}

func init() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://127.0.0.1:4222"
	}
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}

	sc = conn
}

func (stream Stream) Publish(types string, event *event.EventParam) error {
	sc.Publish(types, []byte(event.EventData))
	sc.Publish("event-store", []byte(event.EventData))
	return nil
}

func (stream Stream) Request(types string, event *event.EventParam) (string, error) {
	resp, err := sc.Request(types, []byte(event.EventData), 10*time.Second)
	if err != nil {
		return "", nil
	}
	sc.Publish("event-store", []byte(event.EventData))
	return string(resp.Data), nil
}
