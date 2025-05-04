package pg

import (
	"database/sql"
	"errors"
	"fmt"

	pginit "github.com/volchkovski/go-practicum-metrics/internal/storage/pg/init"

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
	if err = pginit.Initialize(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return &Pg{db}, nil
}

func (pg *Pg) Ping() error {
	return pg.db.Ping()
}

func (pg *Pg) Close() error {
	return pg.db.Close()
}

func (pg *Pg) WriteGauge(name string, value float64) error {
	query := `
	INSERT INTO gauges (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name)
	DO UPDATE SET value = EXCLUDED.value;
	`
	_, err := pg.db.Exec(query, name, value)
	return err
}

func (pg *Pg) WriteCounter(name string, value int64) error {
	query := `
	INSERT INTO counters (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name)
	DO UPDATE SET value = counters.value + EXCLUDED.value;
	`
	_, err := pg.db.Exec(query, name, value)
	return err
}

func (pg *Pg) ReadGauge(name string) (float64, error) {
	query := `SELECT value FROM gauges WHERE name = '$1';`
	var v float64
	if err := pg.db.QueryRow(query, name).Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s not found", name)
		}
		return 0, fmt.Errorf("failed to read gauge %s: %w", name, err)
	}
	return v, nil
}

func (pg *Pg) ReadCounter(name string) (int64, error) {
	query := `SELECT value FROM counters WHERE name = '$1';`
	var v int64
	if err := pg.db.QueryRow(query, name).Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("not found %s", name)
		}
		return 0, fmt.Errorf("failed to read counter %s: %w", name, err)
	}
	return v, nil
}

func (pg *Pg) ReadAllGauges() (gauges map[string]float64, err error) {
	query := `SELECT name, value FROM gauges;`

	rows, err := pg.db.Query(query)
	if err != nil {
		return
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	for rows.Next() {
		var (
			nm string
			v  float64
		)
		if err = rows.Scan(&nm, &v); err != nil {
			return
		}
		gauges[nm] = v
	}

	if err = rows.Err(); err != nil {
		return
	}
	return
}

func (pg *Pg) ReadAllCounters() (counters map[string]int64, err error) {
	query := `SELECT name, value FROM counters;`

	rows, err := pg.db.Query(query)
	if err != nil {
		return
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	for rows.Next() {
		var (
			nm string
			v  int64
		)
		if err = rows.Scan(&nm, &v); err != nil {
			return
		}
		counters[nm] = v
	}

	if err = rows.Err(); err != nil {
		return
	}
	return
}
