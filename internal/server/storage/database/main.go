package database

import (
	"database/sql"
)

func InitSchema(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}

	// server_init.sql
	sqlQuery := `CREATE TABLE IF NOT EXISTS public.counter
		(    
		    "name" varchar NOT NULL,    
		    "value"  bigint  NOT NULL DEFAULT 0,    
		    CONSTRAINT counter_pk PRIMARY KEY ("name")
		);

		CREATE TABLE IF NOT EXISTS public.gauge(    
		    "name" varchar          NOT NULL,    
		    "value" double precision NULL,    
		    CONSTRAINT gauge_pk PRIMARY KEY ("name")
		);`
	_, err := db.Exec(sqlQuery)
	return err
}
