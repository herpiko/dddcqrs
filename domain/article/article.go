package article

import (
	"errors"

	dddcqrs "github.com/herpiko/dddcqrs"
)

// Article domain is defined in proto/article.proto

var ErrInvalidArticleID = errors.New("invalid-article-id")
var ErrInvalidArticleTitle = errors.New("invalid-article-title")
var ErrInvalidArticleBody = errors.New("invalid-article-body")
var ErrInvalidArticleAuthor = errors.New("invalid-article-author")

// Could be used as validator
func NewArticle(title string, body string, author string) (*dddcqrs.Article, error) {
	if title == "" {
		return &dddcqrs.Article{}, ErrInvalidArticleTitle
	}

	if body == "" {
		return &dddcqrs.Article{}, ErrInvalidArticleBody
	}

	if author == "" {
		return &dddcqrs.Article{}, ErrInvalidArticleAuthor
	}

	item := dddcqrs.Article{
		Author: author,
		Title:  title,
		Body:   body,
	}

	return &item, nil
}
