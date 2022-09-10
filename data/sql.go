package data

import (
	"RestApi/schemas"
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgtype"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

const (
	categoryStr = "CATEGORY"
	offerStr    = "OFFER"
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
	}

	if !tables["history"] {
		_, err = db.Exec(
			`create table history
				(
					id         uuid not null,
					"parentId" uuid,
					name varchar not null,
					price      integer,
					time timestamp not null
				);`)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func ValidateImport(parents, offers, categories map[string]bool) bool {
	stmt, err := db.Prepare("select exists(select from item where (id=$1 and type=$2))")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	for parent := range parents {
		var res bool
		err = stmt.QueryRow(parent, "CATEGORY").Scan(&res)
		if !res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}

	for offer := range offers {
		var res bool
		err = stmt.QueryRow(offer, "CATEGORY").Scan(&res)
		if res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}

	for category := range categories {
		var res bool
		err = stmt.QueryRow(category, "OFFER").Scan(&res)
		if res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}

	return true
}

func Import(request *schemas.ImportRequest) error {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()

	stmtItem, err := tx.Prepare(
		`insert into item values($1, $2, $3, $4, $5, $6)
					on conflict (id) do update set 
					"parentId"=$2, "name"=$3, "price"=$4, "time"=$6;`)
	if err != nil {
		log.Println(err)
		return err
	}

	stmtHistory, err := tx.Prepare(`insert into history 
									values($1, $2, $3, $4, $5)`)
	if err != nil {
		log.Println(err)
		return err
	}

	closeStatements := func() {
		stmtItem.Close()
		stmtHistory.Close()
	}

	for _, item := range request.Items {
		var price sql.NullInt64
		var parentId sql.NullString

		id := item.Id
		name := item.Name
		Type := item.Type

		if Type == categoryStr {
			price = sql.NullInt64{}
		} else {
			price = sql.NullInt64{int64(item.Price), true}
		}

		if len(item.ParentId) == 0 {
			parentId = sql.NullString{}
		} else {
			parentId = sql.NullString{item.ParentId, true}
		}

		if _, err = stmtItem.Exec(id, parentId, name, price, Type, request.UpdateDate); err != nil {
			log.Println(err)
			closeStatements()
			return err
		}
		if _, err = stmtHistory.Exec(id, parentId, name, price, request.UpdateDate); err != nil {
			log.Println(err)
			closeStatements()
			return err
		}
	}
	closeStatements()
	return tx.Commit()
}

func Delete(id_ string) error {
	var exist bool
	id := new(pgtype.UUID)
	id.Set(id_)
	err := db.QueryRow(`select exists(select from item where id=$1)`, id).Scan(&exist)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("did not found")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmtFind, err := db.Prepare(`select id from item where (type='CATEGORY' and "parentId"=$1)`)
	if err != nil {
		return err
	}

	stmtDelItem, err := db.Prepare(`delete from item where ("parentId" = $1 or id = $1)`)
	if err != nil {
		return err
	}

	closeStatements := func() {
		stmtDelItem.Close()
		stmtFind.Close()
	}

	queue := list.New()
	queue.PushBack(id)

	for queue.Len() > 0 {
		idDel := queue.Remove(queue.Front()).(*pgtype.UUID)
		rows, err := stmtFind.Query(idDel)
		if err != nil {
			closeStatements()
			return err
		}
		for rows.Next() {
			rows.Scan(&id)
			queue.PushBack(id)
		}

		if err = rows.Err(); err != nil {
			closeStatements()
			return err
		}
		rows.Close()

		if _, err = stmtDelItem.Exec(idDel); err != nil {
			closeStatements()
			return err
		}
	}

	return tx.Commit()
}
