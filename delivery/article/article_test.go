package article_service

import (
	"testing"

	_ "github.com/herpiko/dddcqrs"
)

func TestArticleCreateArticle(t *testing.T) {
	/*
		app := dddcqrs.NewApp()
		util.MigrateClean(app.DB)
		defer app.DB.Close()
		as, err := NewArticleService(
			ArticleRepoPsql(app.DB),
		)
		assert.Equal(t, nil, err)
		as.HTTPHandler(app.Router)
		articleItem, err := article.NewArticle("a", "b", "c")
		assert.Equal(t, nil, err)
		jsonBytes, err := json.Marshal(articleItem)
		assert.Equal(t, nil, err)
		req, _ := http.NewRequest("POST", "/api/articles", string(jsonBytes))
		response := httptest.NewRecorder()
		app.Router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
	*/
}
