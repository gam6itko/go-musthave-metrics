package database

import (
	"database/sql"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
)

// Storage decorator on file.Storage
type Storage struct {
	inner storage.Storage
	db    *sql.DB
}

func NewStorage(inner storage.Storage, db *sql.DB) *Storage {
	return &Storage{
		inner,
		db,
	}
}

func (ths Storage) GaugeSet(name string, val float64) {
	err := ths.gaugeSet(name, val)
	if err == nil {
		return
	}
	ths.inner.GaugeSet(name, val)
}

func (ths Storage) GaugeGet(name string) (float64, bool) {
	result, err := ths.gaugeGet(name)
	if err == nil {
		return result, true
	}
	return ths.inner.GaugeGet(name)
}

func (ths Storage) GaugeAll() map[string]float64 {
	result, err := ths.gaugeAll()
	if err == nil {
		return result
	}
	return ths.inner.GaugeAll()
}

func (ths Storage) CounterInc(name string, val int64) {
	err := ths.counterInc(name, val)
	if err == nil {
		return
	}
	ths.inner.GaugeAll()
}

func (ths Storage) CounterGet(name string) (int64, bool) {
	result, err := ths.counterGet(name)
	if err == nil {
		return result, true
	}
	return ths.inner.CounterGet(name)
}

func (ths Storage) CounterAll() map[string]int64 {
	if result, err := ths.counterAll(); err == nil {
		return result
	}
	return ths.inner.CounterAll()
}

func (ths Storage) gaugeSet(name string, val float64) error {
	_, err := ths.db.Exec("UPDATE `gauge` SET `value` = $1 WHERE `key` = $2", val, name)
	return err
}

func (ths Storage) gaugeGet(name string) (float64, error) {
	row := ths.db.QueryRow("SELECT `value` FROM `gauge` WHERE `key` = $1", name)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var result float64
	if err := row.Scan(result); err != nil {
		return 0, err
	}

	return result, nil
}

func (ths Storage) gaugeAll() (map[string]float64, error) {
	result := make(map[string]float64)

	rows, err := ths.db.Query("SELECT `key`, `value` FROM `gauge`")
	if err != nil {
		return result, err
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var val float64
		if err := rows.Scan(&name, &val); err != nil {
			return result, err
		}

		result[name] = val
	}

	return result, nil
}

func (ths Storage) counterInc(name string, val int64) error {
	_, err := ths.db.Exec("UPDATE `counter` SET `value` = $1 WHERE `key` = $2", val, name)
	return err
}

func (ths Storage) counterGet(name string) (int64, error) {
	row := ths.db.QueryRow("SELECT `value` FROM `counter` WHERE `key` = $1", name)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var result int64
	if err := row.Scan(result); err != nil {
		return 0, err
	}

	return result, nil
}

func (ths Storage) counterAll() (map[string]int64, error) {
	result := make(map[string]int64)

	rows, err := ths.db.Query("SELECT `key`, `value` FROM `counter`")
	if err != nil {
		return result, err
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var val int64
		if err := rows.Scan(&name, &val); err != nil {
			return result, err
		}

		result[name] = val
	}

	return result, nil
}
