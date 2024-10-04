package yps

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type YPSDatabase struct {
	pool *pgxpool.Pool
}

var TheDb *YPSDatabase

func OpenDatabase(connectionUrl string) error {
	pool, err := pgxpool.New(context.Background(), connectionUrl)
	if err != nil {
		return err
	}

	TheDb = &YPSDatabase{
		pool,
	}

	return nil
}

func GetDatabase() *YPSDatabase {
	return TheDb
}

func (db *YPSDatabase) Close() {
	db.pool.Close()
}

func (db *YPSDatabase) GetPage(id string) (page *Page, err error) {
	var rawPage Page

	err = db.pool.QueryRow(context.Background(), "select content, google_form_id, updated_at from pages where id=$1", id).Scan(&rawPage.Content, &rawPage.GoogleFormID, &rawPage.Updated)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return nil, err
	}

	return &rawPage, nil
}
