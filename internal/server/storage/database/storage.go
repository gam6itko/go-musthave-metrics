package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/retrible"
	"github.com/jackc/pgx/v5/pgconn"
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

func (ths Storage) GaugeSet(ctx context.Context, name string, val float64) error {
	query := `INSERT INTO "gauge" ("name", "value") 
		VALUES ($1, $2) 
		ON CONFLICT ("name") DO UPDATE SET value = EXCLUDED.value`
	_, err := ths.db.ExecContext(ctx, query, name, val)
	if ths.isRetrible(err) {
		return retrible.NewError(err)
	}
	return err
}

func (ths Storage) GaugeGet(ctx context.Context, name string) (float64, error) {
	row := ths.db.QueryRowContext(ctx, `SELECT "value" FROM "gauge" WHERE "name" = $1`, name)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var result float64
	if err := row.Scan(&result); err != nil {
		if ths.isRetrible(err) {
			return result, retrible.NewError(err)
		}
		return 0, err
	}

	return result, nil
}

func (ths Storage) GaugeAll(ctx context.Context) (map[string]float64, error) {
	result := make(map[string]float64)

	rows, err := ths.db.QueryContext(ctx, `SELECT "name", value FROM "gauge"`)
	if err != nil {
		if ths.isRetrible(err) {
			return result, retrible.NewError(err)
		}
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

	if err := rows.Err(); err != nil {
		if ths.isRetrible(err) {
			return result, retrible.NewError(err)
		}
		return result, err
	}

	return result, nil
}

func (ths Storage) CounterInc(ctx context.Context, name string, val int64) error {
	query := `INSERT INTO "counter" ("name", "value")
		VALUES ($1, $2)
		ON CONFLICT ("name") DO UPDATE SET "value" = "counter"."value" + EXCLUDED.value`
	_, err := ths.db.ExecContext(ctx, query, name, val)
	if ths.isRetrible(err) {
		return retrible.NewError(err)
	}
	return err
}

func (ths Storage) CounterGet(ctx context.Context, name string) (int64, error) {
	row := ths.db.QueryRowContext(ctx, `SELECT "value" FROM "counter" WHERE "name" = $1`, name)
	if err := row.Err(); err != nil {
		if ths.isRetrible(err) {
			return 0, retrible.NewError(err)
		}
		return 0, err
	}

	var result int64
	if err := row.Scan(&result); err != nil {
		if ths.isRetrible(err) {
			return 0, retrible.NewError(err)
		}
		return 0, err
	}

	return result, nil
}

func (ths Storage) CounterAll(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64)

	rows, err := ths.db.QueryContext(ctx, `SELECT "name", "value" FROM "counter"`)
	if err != nil {
		if ths.isRetrible(err) {
			return result, retrible.NewError(err)
		}
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

	if err := rows.Err(); err != nil {
		if ths.isRetrible(err) {
			return result, retrible.NewError(err)
		}
		return result, err
	}

	return result, nil
}

func (ths Storage) isRetrible(err error) bool {
	switch true {
	case errors.Is(err, &pgconn.ConnectError{}):
		return true
		//todo добавить еще ошибок
	default:
		return false

	}
}
