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
	"github.com/herpiko/dddcqrs/conn/nats"
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

func (s *server) GetEvent(ctx context.Context, eventData *event.EventParam) (*event.EventResponse, error) {
	log.Println("service GetEvent")
	return &event.EventResponse{}, nil
}

func (s *server) CreateEvent(ctx context.Context, eventData *event.EventParam) (*event.ResponseParam, error) {
	log.Println("service CreateEvent")
	respond, err := s.stream.Request(eventData.Channel, eventData)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if respond != "" {
		// Something is wrong, flag the event as fail
		log.Println("flag the event as fail")
		// TODO update eventStore
		return &event.ResponseParam{}, errors.New(respond)
	}
	return &event.ResponseParam{}, nil
}

func (s *server) ListEvent(ctx context.Context, eventData *event.EventParam) (*event.EventResponse, error) {
	log.Println("service ListEvents")
	respond, err := s.stream.Request(eventData.Channel, eventData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result := &event.EventResponse{
		EventId:       eventData.EventId,
		EventType:     eventData.EventType,
		AggregateId:   eventData.AggregateId,
		AggregateType: eventData.AggregateType,
		Channel:       eventData.Channel,
		EventData:     respond,
	}
	return result, err
}
