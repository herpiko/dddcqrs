package psql

import (
	"context"
	"database/sql"
	"log"

	dddcqrs "github.com/herpiko/dddcqrs"
)

type PsqlRepo struct {
	db *sql.DB
}

func New(ctx context.Context, db *sql.DB) (*PsqlRepo, error) {
	return &PsqlRepo{
		db: db,
	}, nil
}

func (pr *PsqlRepo) Create(item *dddcqrs.Article) error {
	_, err := pr.db.Exec(`
    INSERT INTO articles (title, body, author)
    VALUES ($1, $2, $3)
  `, item.Title, item.Body, item.Author)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
