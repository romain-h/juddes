package models

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

func LoadDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("PG_URL"))
	if err != nil {
		return nil, err
	}

	return db, nil
}
