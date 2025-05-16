package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/volchkovski/go-practicum-metrics/internal/storage/pg/migrator"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	MaxOpenConns = 5
	MaxIdleConns
	MaxLifetime = 5 * time.Minute
	MaxIdleTime = 10 * time.Minute
)

type Pg struct {
	db *sql.DB
}

func configureConnPool(db *sql.DB) {
	db.SetMaxOpenConns(MaxOpenConns)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetConnMaxLifetime(MaxLifetime)
	db.SetConnMaxIdleTime(MaxIdleTime)
}

func New(dsn string) (*Pg, error) {
	if err := migrator.Run(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	configureConnPool(db)
	if err := loadQueries(); err != nil {
		return nil, fmt.Errorf("failed to load queries: %w", err)
	}
	return &Pg{db}, nil
}

func (pg *Pg) Ping(ctx context.Context) error {
	return pg.db.PingContext(ctx)
}

func (pg *Pg) Close() error {
	return pg.db.Close()
}

func (pg *Pg) WriteGauge(ctx context.Context, name string, value float64) error {
	_, err := pg.db.ExecContext(ctx, q.InsertGauge, name, value)
	return err
}

func (pg *Pg) WriteCounter(ctx context.Context, name string, value int64) error {
	_, err := pg.db.ExecContext(ctx, q.InsertCounter, name, value)
	return err
}

func (pg *Pg) ReadGauge(ctx context.Context, name string) (float64, error) {
	var val float64
	err := pg.db.QueryRowContext(ctx, q.SelectGaugeValue, name).Scan(&val)
	return val, err
}

func (pg *Pg) ReadCounter(ctx context.Context, name string) (int64, error) {
	var val int64
	err := pg.db.QueryRowContext(ctx, q.SelectCounterValue, name).Scan(&val)
	return val, err
}

func (pg *Pg) ReadAllGauges(ctx context.Context) (gauges map[string]float64, err error) {
	var rows *sql.Rows
	rows, err = pg.db.QueryContext(ctx, q.SelectGauges)
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

func (pg *Pg) ReadAllCounters(ctx context.Context) (counters map[string]int64, err error) {
	var rows *sql.Rows
	rows, err = pg.db.QueryContext(ctx, q.SelectCounters)
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

func (pg *Pg) WriteGaugesCounters(ctx context.Context, gauges map[string]float64, counters map[string]int64) (err error) {
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

	ggStmt, err := tx.PrepareContext(ctx, q.InsertGauge)
	if err != nil {
		return
	}
	for nm, v := range gauges {
		if _, err = ggStmt.Exec(nm, v); err != nil {
			return
		}
	}

	cntStmt, err := tx.PrepareContext(ctx, q.InsertCounter)
	if err != nil {
		return
	}

	for nm, v := range counters {
		if _, err = cntStmt.Exec(nm, v); err != nil {
			return
		}
	}

	err = tx.Commit()
	return
}
