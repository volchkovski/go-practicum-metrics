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
	if err = loadQueries(); err != nil {
		return nil, fmt.Errorf("failed to load queries: %w", err)
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
	_, err := pg.db.Exec(q.InsertGauge, name, value)
	return err
}

func (pg *Pg) WriteCounter(name string, value int64) error {
	_, err := pg.db.Exec(q.InsertCounter, name, value)
	return err
}

func (pg *Pg) ReadGauge(name string) (float64, error) {
	var v float64
	if err := pg.db.QueryRow(q.SelectGaugeValue, name).Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s not found", name)
		}
		return 0, fmt.Errorf("failed to read gauge %s: %w", name, err)
	}
	return v, nil
}

func (pg *Pg) ReadCounter(name string) (int64, error) {
	var v int64
	if err := pg.db.QueryRow(q.SelectCounterValue, name).Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("not found %s", name)
		}
		return 0, fmt.Errorf("failed to read counter %s: %w", name, err)
	}
	return v, nil
}

func (pg *Pg) ReadAllGauges() (gauges map[string]float64, err error) {
	gauges = make(map[string]float64)
	rows, err := pg.db.Query(q.SelectGauges)
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
	counters = make(map[string]int64)
	rows, err := pg.db.Query(q.SelectCounters)
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

func (pg *Pg) WriteGaugesCounters(gauges map[string]float64, counters map[string]int64) (err error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()

	ggStmt, err := tx.Prepare(q.InsertGauge)
	if err != nil {
		return
	}
	for nm, v := range gauges {
		if _, err = ggStmt.Exec(nm, v); err != nil {
			return
		}
	}

	cntStmt, err := tx.Prepare(q.InsertCounter)
	if err != nil {
		return
	}

	for nm, v := range counters {
		if _, err = cntStmt.Exec(nm, v); err != nil {
			return
		}
	}

	return tx.Commit()
}
