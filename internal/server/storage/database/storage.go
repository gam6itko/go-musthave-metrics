package database

import (
	"database/sql"
)

// Storage decorator on file.Storage
type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db,
	}
}

func (ths Storage) GaugeSet(name string, val float64) error {
	query := `INSERT INTO "gauge" ("name", "value") 
		VALUES ($1, $2) 
		ON CONFLICT ("name") DO UPDATE SET value = EXCLUDED.value`
	_, err := ths.db.Exec(query, name, val)
	return err
}

func (ths Storage) GaugeGet(name string) (float64, error) {
	row := ths.db.QueryRow(`SELECT "value" FROM "gauge" WHERE "name" = $1`, name)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var result float64
	if err := row.Scan(&result); err != nil {
		return 0, err
	}

	return result, nil
}

func (ths Storage) GaugeAll() (map[string]float64, error) {
	result := make(map[string]float64)

	rows, err := ths.db.Query(`SELECT "name", value FROM "gauge"`)
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

	if rows.Err() != nil {
		return result, rows.Err()
	}

	return result, nil
}

func (ths Storage) CounterInc(name string, val int64) error {
	query := `INSERT INTO "counter" ("name", "value")
		VALUES ($1, $2)
		ON CONFLICT ("name") DO UPDATE SET "value" = "counter"."value" + EXCLUDED.value`
	_, err := ths.db.Exec(query, name, val)
	return err
}

func (ths Storage) CounterGet(name string) (int64, error) {
	row := ths.db.QueryRow(`SELECT "value" FROM "counter" WHERE "name" = $1`, name)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var result int64
	if err := row.Scan(&result); err != nil {
		return 0, err
	}

	return result, nil
}

func (ths Storage) CounterAll() (map[string]int64, error) {
	result := make(map[string]int64)

	rows, err := ths.db.Query(`SELECT "name", "value" FROM "counter"`)
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

	if rows.Err() != nil {
		return result, rows.Err()
	}

	return result, nil
}
