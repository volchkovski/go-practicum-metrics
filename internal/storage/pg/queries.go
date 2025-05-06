package pg

import (
	"embed"
	"fmt"
	"sync"
)

var (
	once sync.Once
	q    queries
)

type queries struct {
	InsertGauge        string
	InsertCounter      string
	SelectGaugeValue   string
	SelectCounterValue string
	SelectGauges       string
	SelectCounters     string
}

//go:embed queries/*.sql
var queryFS embed.FS

func loadQueries() error {
	var initErr error
	once.Do(func() {
		insertGaugeQ, err := loadQuery("insert_gauge")
		if err != nil {
			initErr = err
			return
		}
		insertCounterQ, err := loadQuery("insert_counter")
		if err != nil {
			initErr = err
			return
		}
		selectGaugeValueQ, err := loadQuery("gauge_value")
		if err != nil {
			initErr = err
			return
		}
		selectCounterValueQ, err := loadQuery("counter_value")
		if err != nil {
			initErr = err
			return
		}
		selectGaugesQ, err := loadQuery("gauges")
		if err != nil {
			initErr = err
			return
		}
		selectCountersQ, err := loadQuery("counters")
		if err != nil {
			initErr = err
			return
		}
		q = queries{
			InsertGauge:        insertGaugeQ,
			InsertCounter:      insertCounterQ,
			SelectGaugeValue:   selectGaugeValueQ,
			SelectCounterValue: selectCounterValueQ,
			SelectGauges:       selectGaugesQ,
			SelectCounters:     selectCountersQ,
		}
	})
	if initErr != nil {
		return fmt.Errorf("failed to load queries: %w", initErr)
	}
	return nil
}

func loadQuery(filename string) (string, error) {
	fp := fmt.Sprintf("queries/%s.sql", filename)
	query, err := queryFS.ReadFile(fp)
	if err != nil {
		return "", fmt.Errorf("failed to read %s query: %w", fp, err)
	}
	return string(query), nil
}
