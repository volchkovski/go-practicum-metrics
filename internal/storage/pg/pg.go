package pg

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Pg struct {
	db *sql.DB
}

func New(dsn string) (*Pg, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Pg{db}, nil
}

func (pg *Pg) Ping() error {
	return pg.db.Ping()
}

func (pg *Pg) Close() error {
	return pg.db.Close()
}
