package database

import (
	"log"
	"os"
	"sync"

	"github.com/olivere/elastic"
)

type ElasticConn struct {
	Conn *elastic.Client
}

var elasticConn *ElasticConn

var lock = &sync.Mutex{}

func NewElasticConn() *ElasticConn {
	lock.Lock()
	defer lock.Unlock()
	if elasticConn == nil {
		elasticConn = &ElasticConn{}
	}
	elasticConn.init()
	return elasticConn
}

func (conn *ElasticConn) init() {

	elasticAddr := os.Getenv("ELASTIC_ADDRESS")
	if elasticAddr == "" {
		elasticAddr = "localhost:9200"
	}

	client, err := elastic.NewClient(
		elastic.SetURL("http://"+elasticAddr),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	elasticConn.Conn = client
}
