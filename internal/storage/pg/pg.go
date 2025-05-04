package pg

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Pg struct {
	db *sql.DB
}

func New(dsn string) (*Pg, error) {
	db, err := sql.Open("pgx", dsn)
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
