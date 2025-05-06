package pginit

import (
	"database/sql"
)

const createTables = `
CREATE TABLE IF NOT EXISTS gauges (
	id    SERIAL PRIMARY KEY,
	name  VARCHAR(255) NOT NULL UNIQUE,
	value DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS counters (
	id    SERIAL PRIMARY KEY,
	name  VARCHAR(255) NOT NULL UNIQUE,
	value BIGINT NOT NULL
);
`

func Initialize(db *sql.DB) error {
	_, err := db.Exec(createTables)
	return err
}
