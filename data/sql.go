package data

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

var db *sql.DB

func StartPG(user, password string) {
	var err error
	db, err = sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@localhost:5432", user, password))
	if err != nil {
		log.Fatal(err)
	}
	fixMissed()
}

func fixMissed() {
	var err error

	var enumExist bool
	err = db.QueryRow(
		`select exists(select from pg_enum 
                       where enumlabel in ('OFFER', 'CATEGORY'))`).Scan(&enumExist)
	if err != nil {
		log.Fatal(err)
	}
	if !enumExist {
		_, err = db.Exec(`create type type as enum ('OFFER', 'CATEGORY')`)
		if err != nil {
			log.Fatal(err)
		}
	}

	tables := map[string]bool{}
	rows, err := db.Query(
		`SELECT tablename FROM pg_catalog.pg_tables 
                 	WHERE schemaname != 'pg_catalog' 
                 		AND schemaname != 'information_schema'`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			log.Fatal(err)
		} else {
			tables[table] = true
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	err = rows.Close()
	if err != nil {
		log.Fatal(err)
	}

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
                     type       varchar not null,
                     time		timestamp not null
                 );`)
		if err != nil {
			log.Fatal(err)
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
					price      integer not null,
					time timestamp not null
				);`)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func AreParents(ids map[string]bool) bool {
	stmt, err := db.Prepare("select exists(select from itema where id=$1)")
	if err != nil {
		log.Println(err)
		return false
	}

	defer stmt.Close()
	for parent, _ := range ids {
		var res bool
		err = stmt.QueryRow(parent).Scan(&parent)
		if !res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}
