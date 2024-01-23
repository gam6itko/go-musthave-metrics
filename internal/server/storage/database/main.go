package database

import (
	"database/sql"
)

func InitSchema(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}

	// server_init.sql
	sql := "CREATE TABLE IF NOT EXISTS public.counter\n(\n    \"name\" varchar NOT NULL,\n    value  bigint  NOT NULL DEFAULT 0,\n    CONSTRAINT counter_pk PRIMARY KEY (\"name\")\n);\n\nCREATE TABLE IF NOT EXISTS public.gauge\n(\n    \"key\" varchar          NOT NULL,\n    value double precision NULL,\n    CONSTRAINT gauge_pk PRIMARY KEY (\"key\")\n);\n"
	_, err := db.Exec(sql)
	return err
}
