package data

import (
	"github.com/gofrs/uuid"
	"log"
)

func ValidateImport(parents, offers, categories map[uuid.NullUUID]bool) bool {
	stmt, err := db.Prepare("select exists(select from item where (id=$1 and type=$2))")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	for parent := range parents {
		var res bool
		err = stmt.QueryRow(parent, categoryStr).Scan(&res)
		if !res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}

	for offer := range offers {
		var res bool
		err = stmt.QueryRow(offer, categoryStr).Scan(&res)
		if res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}

	for category := range categories {
		var res bool
		err = stmt.QueryRow(category, offerStr).Scan(&res)
		if res {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
	}

	return true
}
