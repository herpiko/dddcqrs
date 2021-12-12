package article

import (
	"errors"

	dddcqrs "github.com/herpiko/dddcqrs"
)

var (
	ErrArticleNotFound    = errors.New("the article was not found in the repository")
	ErrFailedToAddArticle = errors.New("failed to add the article to the repository")
)

// Combine to commnad/query funcs and different databases in one repo
// for easier maintenance
type ArticleRepository interface {
	// Commands
	Create(*dddcqrs.Article) error
	CreateAggregate(*dddcqrs.ArticleAggregateRoot) error

	// Query
	List(*dddcqrs.Articles) (*dddcqrs.Articles, error)
}
