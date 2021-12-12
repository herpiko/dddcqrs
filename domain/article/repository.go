package article

import (
	dddcqrs "github.com/herpiko/dddcqrs"
)

// Combine to commnad/query funcs and different databases in one repo
// for easier maintenance. We already have distinct executables for C/Q.
type ArticleRepository interface {
	// Commands
	Create(*dddcqrs.Article) (int32, error)
	CreateAggregate(*dddcqrs.ArticleAggregateRoot) error

	// Query
	List(*dddcqrs.Articles) (*dddcqrs.Articles, error)
	Get(id string) (*dddcqrs.ArticleAggregateRoot, error)
}
