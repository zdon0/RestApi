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

	var enumExist bool

	err = db.QueryRow(
		`select exists(select 1 from pg_enum 
                       where (enumlabel='OFFER' or enumlabel='CATEGORY'))`).Scan(&enumExist)

	if err != nil {
		log.Fatal(err)
	}

	if !enumExist {
		_, err = db.Exec(`create type type as enum ('OFFER', 'CATEGORY')`)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	tables := map[string]bool{}
	rows, err := db.Query(
		`SELECT tablename FROM pg_catalog.pg_tables 
                 	WHERE schemaname != 'pg_catalog' 
                 		AND schemaname != 'information_schema'`)
	if err != nil {
		log.Fatal(err.Error())
	}
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			tables[table] = true
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err.Error())
	}
	rows.Close()

	if !tables["item"] {
		_, err = db.Exec(
			`create table item
                 (
                     id         uuid    not null
                         constraint id
                             primary key,
                     "parentId" uuid,
                     name       varchar not null,
                     price      integer,
                     type       varchar,
                     time		timestamp
                 );`)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if !tables["history"] {
		_, err = db.Exec(
			`create table history
				(
					id         uuid not null
						constraint "historyId"
							references item (id),
					"parentId" uuid,
					price      integer,
					time timestamp
				);`)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	return nil
}
