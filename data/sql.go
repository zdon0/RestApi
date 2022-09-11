package data

import (
	"RestApi/schemas"
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"strings"
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

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	rows.Close()

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
					id         uuid not null
						constraint "historyId"
            			references item (id)
            			on delete cascade,
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

func ValidateImport(parents, offers, categories map[uuid.NullUUID]bool) bool {
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

	stmtHistory, err := tx.Prepare(`insert into history values($1, $2, $3, $4, $5)`)
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
		id := item.Id
		name := item.Name
		Type := item.Type
		parentId := item.ParentId

		if Type == categoryStr {
			price = sql.NullInt64{}
		} else {
			price = sql.NullInt64{int64(item.Price), true}
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

func Delete(id string) error {

	if !IsExist(id) {
		return errors.New("not found")
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

	queue := list.New()
	queue.PushBack(id)
	deleteArray := make([]any, 0, 20)

	for queue.Len() > 0 {
		idDel := queue.Remove(queue.Front()).(string)
		deleteArray = append(deleteArray, idDel)

		rows, err := stmtFind.Query(idDel)
		if err != nil {
			stmtFind.Close()
			return err
		}
		for rows.Next() {
			rows.Scan(&id)
			queue.PushBack(id)
		}
		if err = rows.Err(); err != nil {
			stmtFind.Close()
			return err
		}
		rows.Close()
	}

	holdersArray := make([]string, len(deleteArray))
	for i := 1; i <= len(holdersArray); i++ {
		holdersArray[i-1] = fmt.Sprintf("$%d", i)
	}

	placeHolders := strings.Join(holdersArray, ",")

	query := fmt.Sprintf(`delete from item where (id in (%s) or "parentId" in (%s))`,
		placeHolders, placeHolders)

	if _, err = db.Exec(query, deleteArray...); err != nil {
		stmtFind.Close()
		return err
	}

	stmtFind.Close()
	return tx.Commit()
}

type nodeItem struct {
	id       string
	parentId sql.NullString
	name     string
	price    sql.NullInt64
	Type     string

	sum      int
	len      int
	parent   *nodeItem
	children []*nodeItem
}

func Nodes(id string) (map[any]any, error) {
	res := map[any]any{}
	stmtItem, err := db.Prepare(`select id, "parentId", name, price, type from item where id=$1`)
	if err != nil {
		return res, err
	}
	defer stmtItem.Close()

	stmtChildren, err := db.Prepare(`select id from item where ("parentId"=$1 and type=$2)`)
	if err != nil {
		return res, err
	}
	defer stmtChildren.Close()

	head := &nodeItem{}

	if err = stmtItem.QueryRow(id).Scan(head.id, head.parentId, head.name, head.price, head.price); err != nil {
		return res, err
	}

	func(parent *nodeItem) {

	}(head)

	return res, nil
}

func IsExist(id string) bool {
	var exist bool

	if err := db.QueryRow(`select exists(select from item where id=$1)`, id).Scan(&exist); err != nil {
		return false
	}
	return exist
}
