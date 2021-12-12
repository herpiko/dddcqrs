package article_service

import (
	"context"
	"database/sql"

	"github.com/herpiko/dddcqrs/domain/article"
	storage "github.com/herpiko/dddcqrs/domain/article/storage"
	"github.com/olivere/elastic"
)

type ArticleDelivery struct {
	Articles article.ArticleRepository
}

type ArticleConfig func(ad *ArticleDelivery) error

func NewArticleDelivery(cfgs ...ArticleConfig) (*ArticleDelivery, error) {
	ad := &ArticleDelivery{}
	for _, cfg := range cfgs {
		err := cfg(ad)
		if err != nil {
			return nil, err
		}
	}
	return ad, nil
}

// Hook article delivery with psql and elastic, should be used for command
func WithPsqlAndElastic(db *sql.DB, el *elastic.Client) ArticleConfig {
	return func(ad *ArticleDelivery) error {
		articleRepo, err := storage.New(context.Background(), db, el)
		if err != nil {
			return err
		}
		ad.Articles = articleRepo
		return nil
	}
}

// Hook article delivery with elastic only, should be used for query
func WithElastic(el *elastic.Client) ArticleConfig {
	return func(ad *ArticleDelivery) error {
		articleRepo, err := storage.New(context.Background(), nil, el)
		if err != nil {
			return err
		}
		ad.Articles = articleRepo
		return nil
	}
}
