package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	pginit "github.com/volchkovski/go-practicum-metrics/internal/storage/pg/init"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var backoffs = []time.Duration{
	0 * time.Second,
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

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

func (pg *Pg) WriteGauge(name string, value float64) (err error) {
	for _, backoff := range backoffs {
		_, err = pg.db.Exec(q.InsertGauge, name, value)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	return
}

func (pg *Pg) WriteCounter(name string, value int64) (err error) {
	for _, backoff := range backoffs {
		_, err = pg.db.Exec(q.InsertCounter, name, value)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	return
}

func (pg *Pg) ReadGauge(name string) (val float64, err error) {
	for _, backoff := range backoffs {
		err = pg.db.QueryRow(q.SelectGaugeValue, name).Scan(&val)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	return
}

func (pg *Pg) ReadCounter(name string) (val int64, err error) {
	for _, backoff := range backoffs {
		err = pg.db.QueryRow(q.SelectCounterValue, name).Scan(&val)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	return
}

func (pg *Pg) ReadAllGauges() (gauges map[string]float64, err error) {
	var rows *sql.Rows
	for _, backoff := range backoffs {
		rows, err = pg.db.Query(q.SelectGauges)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	if err != nil {
		return
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	gauges = make(map[string]float64)
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

	err = rows.Err()
	return
}

func (pg *Pg) ReadAllCounters() (counters map[string]int64, err error) {
	var rows *sql.Rows
	for _, backoff := range backoffs {
		rows, err = pg.db.Query(q.SelectCounters)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	if err != nil {
		return
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	counters = make(map[string]int64)
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

	err = rows.Err()
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

	for _, backoff := range backoffs {
		err = tx.Commit()
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			time.Sleep(backoff)
			continue
		}
		break
	}
	return
}
