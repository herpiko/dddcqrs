package main

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	dddcqrs "github.com/herpiko/dddcqrs"
	event "github.com/herpiko/dddcqrs"
	"github.com/herpiko/dddcqrs/nats"
)

type server struct {
	//ev     db.EventStore
	stream nats.Stream
	//query  db.Query
}

func main() {
	port := ":4040"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	event.RegisterEventStoreServer(srv, &server{})
	event.RegisterAddServiceServer(srv, &server{})
	reflection.Register(srv)

	log.Println("hub running at " + port)
	if e := srv.Serve(listener); e != nil {
		panic(e)
	}
}

func (s *server) CreateArticle(ctx context.Context, request *event.Article) (*event.Article, error) {
	log.Println("service CreateArticle")
	return &event.Article{}, nil
}

func (s *server) ListArticle(ctx context.Context, request *event.ListArticleParam) (*event.Articles, error) {
	log.Println("service ListArticle")
	listArticle := dddcqrs.Articles{}
	return &listArticle, nil
}

func (s *server) GetEvents(ctx context.Context, eventData *event.EventFilter) (*event.EventResponse, error) {
	log.Println("service GetEvents")
	return &event.EventResponse{}, nil
}

func (s *server) CreateEvent(ctx context.Context, eventData *event.EventParam) (*event.ResponseParam, error) {
	log.Println("service CreateEvents")
	// TODO event store
	//createEvent := s.ev.CreateEvent(eventData)

	// Publish to nats
	// go s.stream.Publish(eventData.channel, eventData)

	// Request to command worker(s)
	respond, err := s.stream.Request(eventData.Channel, eventData)
	log.Println(err)
	log.Println(respond)
	if respond != "" {
		// Something is wrong, flag the event as fail
		log.Println("flag the event as fail")
		// TODO update eventStore
		return &event.ResponseParam{}, errors.New(respond)
	}

	//fmt.Println(createEvent)
	//if createEvent == nil {
	//		return &event.ResponseParam{}, errors.Wrap(createEvent, "error from RPC server")
	//}
	return &event.ResponseParam{}, nil
}
