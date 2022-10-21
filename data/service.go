package data

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

var db *sql.DB

func StartPG(user, password string) {
	var err error
	db, err = sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@database:5432/postgres?sslmode=disable",
		user, password))
	if err != nil {
		log.Fatal(err)
	}
	fixMissed()
}

func fixMissed() {
	var err error

	_, err = db.Exec(`drop type if exists type; create type type as enum ('OFFER', 'CATEGORY');`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`create table if not exists item
                 (
                     id         uuid    not null
                         constraint "itemId"
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

	_, err = db.Exec(
		`create table if not exists price_history
				(
					id         uuid not null
						constraint "historyId"
            			references item (id)
            			on delete cascade,
					price      integer,
					time timestamp not null
				);`)

	if err != nil {
		log.Fatal(err)
	}
}

func isExist(id string) bool {
	var exist bool

	if err := db.QueryRow(`select exists(select from item where id=$1)`, id).Scan(&exist); err != nil {
		return false
	}
	return exist
}

func generatePlaceHolders(size, start int) string {
	holdersArray := make([]string, size)
	for i := 1; i <= size; i++ {
		holdersArray[i-1] = fmt.Sprintf("$%d", i+start)
	}
	return strings.Join(holdersArray, ",")
}
