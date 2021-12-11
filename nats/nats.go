package nats

import (
	"time"

	event "github.com/herpiko/dddcqrs"
	"github.com/nats-io/nats.go"
)

type Stream struct{}

var sc *nats.Conn

func init() {
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	sc = conn
}

func (stream Stream) Publish(types string, event *event.EventParam) error {
	// Simple Synchronous Publisher

	//publish to service receiver
	sc.Publish(types, []byte(event.EventData))

	//publish to log
	sc.Publish("log", []byte(event.EventData))
	return nil
}

func (stream Stream) Request(types string, event *event.EventParam) (string, error) {
	// Simple Synchronous Publisher

	//publish to service receiver
	resp, err := sc.Request(types, []byte(event.EventData), 10*time.Second)
	if err != nil {
		return "", nil
	}

	//publish to log
	sc.Publish("log", []byte(event.EventData))
	return string(resp.Data), nil
}
