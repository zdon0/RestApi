package data

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

var db *sql.DB

func Start(user, password string) error {
	var err error
	db, err = sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@localhost:5432", user, password))
	if err != nil {
		return err
	}
	rows, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables " +
		"WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
