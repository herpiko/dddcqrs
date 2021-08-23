package article_service

import (
	"context"
	"database/sql"

	"github.com/herpiko/dddcqrs/domain/article"
	"github.com/herpiko/dddcqrs/domain/article/psql"
)

type ArticleDelivery struct {
	articles article.ArticleRepository
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

func ArticleRepoWithPsql(db *sql.DB) ArticleConfig {
	return func(ad *ArticleDelivery) error {
		articleRepo, err := psql.New(context.Background(), db)
		if err != nil {
			return err
		}
		ad.articles = articleRepo
		return nil
	}
}
