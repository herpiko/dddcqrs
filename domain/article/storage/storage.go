package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	dddcqrs "github.com/herpiko/dddcqrs"
	erq "github.com/herpiko/dddcqrs/internal/query"
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

func (as *ArticleStorage) Create(item *dddcqrs.Article) (int32, error) {
	var id int32
	err := as.psql.QueryRow(`
    INSERT INTO articles (title, body, author, created_at)
    VALUES ($1, $2, $3, $4) RETURNING id
  `, item.Title, item.Body, item.Author, time.Unix(0, item.CreatedAt)).
		Scan(&id)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return id, nil
}

func (as *ArticleStorage) CreateAggregate(item *dddcqrs.ArticleAggregateRoot) error {

	ctx := context.Background()
	jsonBytes, _ := json.Marshal(item)
	_, err := as.el.Index().
		Index("article").
		Id(item.AggregateId).
		BodyJson(string(jsonBytes)).
		Type("fulltext").
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

		// Pagination and order
	from := 0
	if articleParam.Page > 1 {
		from = int((articleParam.Page * articleParam.Limit) - articleParam.Limit)
	}

	// olivere/elastic does not support multi fuzzy query yet,
	// use raw query instead.
	rawQuery := erq.ElRawQuery{
		Sort: []erq.SortItem{},
		From: strconv.Itoa(from),
		Size: strconv.Itoa(int(articleParam.Limit)),
	}
	rawQuery.Query = nil
	rawQuery.Sort = append(rawQuery.Sort, erq.SortItem{
		CreatedAt: erq.SortOpt{
			Order: "desc",
		},
	},
	)

	// Article search
	if articleParam.ArticleFilter != "" {
		rawQuery.Query = &erq.Query{
			MultiMatch: erq.MultiMatch{
				Fields:    []string{},
				Query:     articleParam.ArticleFilter,
				Fuzziness: "AUTO",
			},
		}
		rawQuery.Query.MultiMatch.Fields = append(rawQuery.Query.MultiMatch.Fields, "title")
		rawQuery.Query.MultiMatch.Fields = append(rawQuery.Query.MultiMatch.Fields, "body")

		// Author search
	} else if articleParam.AuthorFilter != "" {
		rawQuery.Query = &erq.Query{
			MultiMatch: erq.MultiMatch{
				Fields:    []string{},
				Query:     articleParam.AuthorFilter,
				Fuzziness: "AUTO",
			},
		}
		rawQuery.Query.MultiMatch.Fields = append(rawQuery.Query.MultiMatch.Fields, "author")
	}
	s, _ := json.Marshal(rawQuery)
	searchService = searchService.Source(string(s))
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

func (as *ArticleStorage) Get(id string) (*dddcqrs.ArticleAggregateRoot, error) {
	ctx := context.Background()
	searchSource := elastic.NewSearchSource()
	searchService := as.el.Search().
		Index("article").
		SearchSource(searchSource)

	// olivere/elastic does not support multi fuzzy query yet,
	// use raw query instead.
	rawQuery := erq.ElRawQuery{
		Sort: []erq.SortItem{},
		From: "0",
		Size: "1",
	}
	rawQuery.Query = nil
	rawQuery.Sort = append(rawQuery.Sort, erq.SortItem{
		CreatedAt: erq.SortOpt{
			Order: "desc",
		},
	},
	)

	rawQuery.Query = &erq.Query{
		MultiMatch: erq.MultiMatch{
			Fields:    []string{},
			Query:     id,
			Fuzziness: "AUTO",
		},
	}
	rawQuery.Query.MultiMatch.Fields = append(rawQuery.Query.MultiMatch.Fields, "id")

	s, _ := json.Marshal(rawQuery)
	log.Println(string(s))
	searchService = searchService.Source(string(s))

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(searchResult.Hits.Hits) < 1 {
		err = errors.New("item-not-found")
		log.Println(err)
		return nil, err
	}

	var articleItem *dddcqrs.ArticleAggregateRoot
	for _, hit := range searchResult.Hits.Hits {
		err := json.Unmarshal(*hit.Source, &articleItem)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(articleItem)
		break
	}
	return articleItem, nil
}
