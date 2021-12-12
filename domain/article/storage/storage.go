package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	dddcqrs "github.com/herpiko/dddcqrs"
	"github.com/olivere/elastic"
)

type ArticleStorage struct {
	psql *sql.DB
	el   *elastic.Client
}

func New(ctx context.Context, db *sql.DB, elasticClient *elastic.Client) (*ArticleStorage, error) {
	return &ArticleStorage{
		psql: db,
		el:   elasticClient,
	}, nil
}

func (as *ArticleStorage) Create(item *dddcqrs.Article) error {
	_, err := as.psql.Exec(`
    INSERT INTO articles (title, body, author)
    VALUES ($1, $2, $3)
  `, item.Title, item.Body, item.Author)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (as *ArticleStorage) CreateAggregate(item *dddcqrs.ArticleAggregateRoot) error {
	log.Println("=======================")
	ctx := context.Background()
	jsonBytes, _ := json.Marshal(item)
	_, err := as.el.Index().
		Index("article").
		Id(item.AggregateId).
		BodyJson(string(jsonBytes)).
		Type("_doc").
		Do(ctx)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (as *ArticleStorage) List(articleParam *dddcqrs.Articles) (*dddcqrs.Articles, error) {
	ctx := context.Background()
	searchSource := elastic.NewSearchSource()

	searchService := as.el.Search().
		Index("article").
		SearchSource(searchSource)

	if articleParam.Search != "" {
		q := elastic.NewMultiMatchQuery(articleParam.Search, "title.keyword", "body.keyword").
			Type("phrase_prefix")
		searchService = searchService.Query(q)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	articleList := &dddcqrs.Articles{}
	for _, hit := range searchResult.Hits.Hits {
		var articleItem *dddcqrs.ArticleAggregateRoot
		err := json.Unmarshal(*hit.Source, &articleItem)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(articleItem)
		articleList.Data = append(articleList.Data, articleItem)
	}
	return articleList, nil
}

/*
func InsertToElastic(deposit *dddcqrs.ArticleAggregateRoot) {
	ctx := context.Background()

	elData, _ := json.Marshal(deposit)
	js := string(elData)
	_, elErr := elasticConn.Index().
		Index(elIndex).
		Id(deposit.AggregateID).
		BodyJson(js).
		Do(ctx)

	if elErr != nil {
		log.Println(elErr)
	}

}
func UpadateToElastic(AggregateID string) {
	ctx := context.Background()

	_, elErr := elasticConn.Update().Index(elIndex).Id(AggregateID).Doc(map[string]interface{}{"approve": 1}).Do(ctx)
	if elErr != nil {
		log.Println(elErr)
	}
}
*/
