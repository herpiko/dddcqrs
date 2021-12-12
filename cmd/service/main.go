package main

import (
	"context"
	"encoding/json"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/herpiko/dddcqrs"
	event "github.com/herpiko/dddcqrs"
	"github.com/herpiko/dddcqrs/conn/nats"
	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
)

type server struct {
	stream nats.Stream
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	port := ":4040"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	event.RegisterEventStoreServer(srv, &server{})
	dddcqrs.RegisterArticleServiceServer(srv, &server{})
	reflection.Register(srv)

	log.Println("grpc hub running at " + port)
	if e := srv.Serve(listener); e != nil {
		panic(e)
	}
}

func (s *server) CreateArticle(ctx context.Context, articleItem *dddcqrs.Article) (*dddcqrs.ArticleId, error) {

	jsonBytes, err := json.Marshal(articleItem)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	evID := uuid.NewV4()
	agID := uuid.NewV4()
	eventItem := &event.EventParam{
		Channel:       "article-created",
		EventType:     "article-created",
		AggregateType: "article",
		EventId:       evID.String(),
		AggregateId:   agID.String(),
		EventData:     string(jsonBytes),
	}
	response, err := s.InvokeEvent(context.Background(), eventItem)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(response)

	return &dddcqrs.ArticleId{Id: response.EventData}, nil
}

func (s *server) ListArticle(ctx context.Context, articleItem *dddcqrs.Articles) (*dddcqrs.ArticleData, error) {
	jsonBytes, err := json.Marshal(articleItem)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(string(jsonBytes))
	evID := uuid.NewV4()
	agID := uuid.NewV4() // Not used yet
	eventItem := &event.EventParam{
		Channel:       "article-list",
		EventType:     "article-list",
		AggregateType: "article",
		EventId:       evID.String(),
		AggregateId:   agID.String(),
		EventData:     string(jsonBytes),
	}
	res, err := s.InvokeEvent(context.Background(), eventItem)
	log.Println(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(res)
	return &dddcqrs.ArticleData{Data: res.EventData}, nil
}

func (s *server) GetArticle(ctx context.Context, articleId *dddcqrs.ArticleId) (*dddcqrs.ArticleData, error) {

	evID := uuid.NewV4()
	agID := uuid.NewV4() // Not used yet
	eventItem := &event.EventParam{
		Channel:       "article-get",
		EventType:     "article-get",
		AggregateType: "article",
		EventId:       evID.String(),
		AggregateId:   agID.String(),
		EventData:     articleId.Id,
	}
	res, err := s.InvokeEvent(context.Background(), eventItem)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &dddcqrs.ArticleData{Data: res.EventData}, nil
}

func (s *server) InvokeEvent(ctx context.Context, eventData *event.EventParam) (*event.EventResponse, error) {
	log.Println("service InvokeEvent")
	respond, err := s.stream.Request(eventData.Channel, eventData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(respond)

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
