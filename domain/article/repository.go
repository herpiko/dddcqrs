package article

import (
	"errors"

	dddcqrs "github.com/herpiko/dddcqrs"
)

var (
	ErrArticleNotFound    = errors.New("the article was not found in the repository")
	ErrFailedToAddArticle = errors.New("failed to add the article to the repository")
)

// Pluggable against any db/storage system
// See psql/psql.go as implementation example
type ArticleRepository interface {
	Create(*dddcqrs.Article) error
}
