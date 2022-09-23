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
	"time"
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

	if !tables["price_history"] {
		_, err = db.Exec(
			`create table price_history
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

func isExist(id string) bool {
	var exist bool

	if err := db.QueryRow(`select exists(select from item where id=$1)`, id).Scan(&exist); err != nil {
		return false
	}
	return exist
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
					"parentId"=$2, name=$3, price=$4, time=$6;`)
	if err != nil {
		log.Println(err)
		return err
	}

	stmtHistory, err := tx.Prepare(
		`insert into price_history select $1, $2, $3
					where not exists(select from price_history
         					where (id=$1 and $2=(select price from price_history
                                      where (id=$1) order by time desc limit 1)))`)
	if err != nil {
		log.Println(err)
		return err
	}

	closeStatements := func() {
		stmtItem.Close()
		stmtHistory.Close()
	}

	parents := map[string]bool{}

	for _, item := range request.Items {
		var price sql.NullInt64
		id := item.Id
		name := item.Name
		Type := item.Type
		parentId := item.ParentId

		if parentId.Valid {
			parents[parentId.UUID.String()] = true
		}

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

		if Type == "OFFER" {
			if _, err = stmtHistory.Exec(id, price, request.UpdateDate); err != nil {
				log.Println(err)
				closeStatements()
				return err
			}
		}
	}
	closeStatements()

	stmtFindParent, err := tx.Prepare(`select "parentId" from item where id=$1`)

	if err != nil {
		return err
	}

	queue := list.New()
	updateArray := make([]any, 0, len(parents))
	for parent := range parents {
		queue.PushBack(parent)
		updateArray = append(updateArray, parent)
	}

	for queue.Len() > 0 {
		parent := queue.Remove(queue.Front())

		rows, err := stmtFindParent.Query(parent)
		if err != nil {
			stmtFindParent.Close()
			return err
		}

		for rows.Next() {
			var id string
			rows.Scan(&id)
			if len(id) > 0 && !parents[id] {
				queue.PushBack(id)
				updateArray = append(updateArray, id)
				parents[id] = true
			}
		}

		if rows.Err() != nil {
			stmtFindParent.Close()
			return rows.Err()
		}

	}

	if len(updateArray) > 0 {
		placeHolders := generatePlaceHolders(len(updateArray), 1)
		queryArgs := []any{request.UpdateDate}
		queryArgs = append(queryArgs, updateArray...)

		query := fmt.Sprintf(`update item set time=$1 where id in (%s)`, placeHolders)

		if _, err = tx.Exec(query, queryArgs...); err != nil {
			stmtFindParent.Close()
			return err
		}
	}
	stmtFindParent.Close()
	return tx.Commit()
}

func Delete(id string) error {

	if !isExist(id) {
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
	}

	placeHolders := generatePlaceHolders(len(deleteArray), 0)

	query := fmt.Sprintf(`delete from item where (id in (%s) or "parentId" in (%s))`,
		placeHolders, placeHolders)

	if _, err = db.Exec(query, deleteArray...); err != nil {
		stmtFind.Close()
		return err
	}

	stmtFind.Close()
	return tx.Commit()
}

func generatePlaceHolders(size, start int) string {
	holdersArray := make([]string, size)
	for i := 1; i <= size; i++ {
		holdersArray[i-1] = fmt.Sprintf("$%d", i+start)
	}
	return strings.Join(holdersArray, ",")
}

func Nodes(id string) (map[string]any, error) {
	response := map[string]any{}

	if !isExist(id) {
		return response, errors.New("not found")
	}

	stmtItem, err := db.Prepare(`select id, "parentId", name, price, type, time from item where id=$1`)
	if err != nil {
		return response, err
	}
	defer stmtItem.Close()

	stmtChildren, err := db.Prepare(`select id from item where ("parentId"=$1 and type=$2)`)
	if err != nil {
		return response, err
	}
	defer stmtChildren.Close()

	head := &nodeItem{}

	if err = stmtItem.QueryRow(id).Scan(&head.id, &head.parentId,
		&head.name, &head.price, &head.Type, &head.date); err != nil {
		return response, err
	}

	if err = head.fill(stmtChildren, stmtItem); err != nil {
		return response, err
	}

	response = head.treeToMap()
	return response, nil
}

func Sales(target time.Time) (map[string][]map[string]any, error) {
	response := map[string][]map[string]any{"items": {}}
	dayAgo := target.Add(-24 * time.Hour)
	rows, err := db.Query(`select distinct i.* from item i 
    								join price_history p on i.id = p.id 
                    					where (p.time >= $1 and p.time <= $2)`,
		dayAgo, target)

	if err != nil {
		return map[string][]map[string]any{}, err
	}

	for rows.Next() {

		item := map[string]any{}
		var id, name, Type string
		var price int
		var parentId sql.NullString
		var date time.Time

		if err = rows.Scan(&id, &parentId, &name, &price, &Type, &date); err != nil {
			return map[string][]map[string]any{}, err
		}

		item["id"] = id
		item["name"] = name
		item["price"] = price
		item["type"] = Type
		item["date"] = date.Format("2006-01-02T15:04:05.000Z")
		if parentId.Valid {
			item["parentId"] = parentId.String
		} else {
			item["parentId"] = nil
		}

		response["items"] = append(response["items"], item)
	}

	if rows.Err() != nil {
		return map[string][]map[string]any{}, err
	}
	return response, nil
}
