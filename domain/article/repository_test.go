package article

import (
	"context"
	"strconv"
	"testing"
	"time"

	dddcqrs "github.com/herpiko/dddcqrs"
	el "github.com/herpiko/dddcqrs/conn/elastic"
	psql "github.com/herpiko/dddcqrs/conn/psql"
	storage "github.com/herpiko/dddcqrs/domain/article/storage"
	"github.com/stretchr/testify/assert"
)

func TestCreateArticle(t *testing.T) {
	db := psql.NewPsqlConn().DB
	defer db.Close()
	elastic := el.NewElasticConn().Conn
	articleRepo, err := storage.New(context.Background(), db, elastic)
	assert.Equal(t, nil, err)

	articleItem := &dddcqrs.Article{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	lastInsertId, err := articleRepo.Create(articleItem)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, lastInsertId > 0)

	err = articleRepo.CreateAggregate(&dddcqrs.ArticleAggregateRoot{
		Id:        strconv.Itoa(int(lastInsertId)),
		Title:     articleItem.Title,
		Body:      articleItem.Body,
		Author:    articleItem.Author,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	assert.Equal(t, nil, err)

}

func TestGetArticle(t *testing.T) {
	db := psql.NewPsqlConn().DB
	defer db.Close()
	elastic := el.NewElasticConn().Conn
	articleRepo, err := storage.New(context.Background(), db, elastic)
	assert.Equal(t, nil, err)

	articleItem := &dddcqrs.Article{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	lastInsertId, err := articleRepo.Create(articleItem)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, lastInsertId > 0)

	err = articleRepo.CreateAggregate(&dddcqrs.ArticleAggregateRoot{
		Id:        strconv.Itoa(int(lastInsertId)),
		Title:     articleItem.Title,
		Body:      articleItem.Body,
		Author:    articleItem.Author,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	assert.Equal(t, nil, err)

	time.Sleep(1000 * time.Millisecond)

	created, err := articleRepo.Get(strconv.Itoa(int(lastInsertId)))
	assert.Equal(t, nil, err)
	assert.Equal(t, created.Title, articleItem.Title)
	assert.Equal(t, created.Body, articleItem.Body)
	assert.Equal(t, created.Author, articleItem.Author)
}

func TestListArticle(t *testing.T) {
	db := psql.NewPsqlConn().DB
	defer db.Close()
	elastic := el.NewElasticConn().Conn
	articleRepo, err := storage.New(context.Background(), db, elastic)
	assert.Equal(t, nil, err)

	articleItem := &dddcqrs.Article{
		Title:  "pygmy",
		Body:   "nicknamed pygmy for his diminutive size",
		Author: "chuck palahniuk",
	}
	lastInsertId, err := articleRepo.Create(articleItem)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, lastInsertId > 0)

	err = articleRepo.CreateAggregate(&dddcqrs.ArticleAggregateRoot{
		Id:        strconv.Itoa(int(lastInsertId)),
		Title:     articleItem.Title,
		Body:      articleItem.Body,
		Author:    articleItem.Author,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	assert.Equal(t, nil, err)

	time.Sleep(1000 * time.Millisecond)

	list, err := articleRepo.List(&dddcqrs.Articles{Page: 1, Limit: 10})
	assert.Equal(t, nil, err)
	assert.Equal(t, true, len(list.Data) > 0)
	assert.Equal(t, list.Data[0].Title, articleItem.Title)
	assert.Equal(t, list.Data[0].Body, articleItem.Body)
	assert.Equal(t, list.Data[0].Author, articleItem.Author)
}
