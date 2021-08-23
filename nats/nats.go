package nats

import (
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
