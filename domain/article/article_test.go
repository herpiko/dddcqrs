package article_test

import (
	"testing"

	"github.com/herpiko/dddcqrs"
	"github.com/herpiko/dddcqrs/domain/article"
)

func TestArticle_NewArticle(t *testing.T) {
	type testCase struct {
		test        string
		article     *dddcqrs.Article
		expectedErr error
	}
	testCases := []testCase{
		{
			test:        "Empty title",
			article:     &dddcqrs.Article{},
			expectedErr: article.ErrInvalidArticleTitle,
		}, {
			test:        "Empty body",
			article:     &dddcqrs.Article{Title: "ok"},
			expectedErr: article.ErrInvalidArticleBody,
		}, {
			test:        "Empty author",
			article:     &dddcqrs.Article{Title: "ok", Body: "ok"},
			expectedErr: article.ErrInvalidArticleAuthor,
		}, {
			test: "Valid payload",
			article: &dddcqrs.Article{Title: "Robohnya Surau Kami",
				Body:   "The quick brown fox jumps over the lazy dog",
				Author: "AA Navis"},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			_, err := article.NewArticle(tc.article.Title, tc.article.Body, tc.article.Author)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
