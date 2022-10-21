package data

import (
	"database/sql"
	"time"
)

type nodeItem struct {
	id, name, Type string
	parentId       sql.NullString
	price          sql.NullInt64
	date           time.Time

	sum, len int64
	parent   *nodeItem
	children []*nodeItem
}

func (n *nodeItem) treeToMap() map[string]any {
	res := map[string]any{}
	res["id"] = n.id
	res["name"] = n.name
	res["type"] = n.Type
	res["date"] = n.date.Format("2006-01-02T15:04:05.000Z")

	if n.parentId.Valid {
		res["parentId"] = n.parentId.String
	} else {
		res["parentId"] = nil
	}

	if res["type"] == offerStr {
		res["price"] = n.price.Int64
		res["children"] = nil
	} else if res["type"] == categoryStr {
		if n.len != 0 {
			res["price"] = n.sum / n.len
		} else {
			res["price"] = 0
		}
		children := make([]map[string]any, len(n.children))
		for i, child := range n.children {
			children[i] = child.treeToMap()
		}
		res["children"] = children
	}
	return res
}

func (n *nodeItem) fill(stmtChildren, stmtItem *sql.Stmt) error {
	childrenOffer, err := stmtChildren.Query(n.id, offerStr)
	if err != nil {
		return err
	}
	defer childrenOffer.Close()

	var total, sum int64

	for childrenOffer.Next() {

		newChildren := &nodeItem{parent: n}

		var childrenId string
		if err = childrenOffer.Scan(&childrenId); err != nil {
			return err
		}

		if err = stmtItem.QueryRow(childrenId).Scan(&newChildren.id, &newChildren.parentId,
			&newChildren.name, &newChildren.price, &newChildren.Type, &newChildren.date); err != nil {
			return err
		}

		total++
		sum += newChildren.price.Int64
		n.children = append(n.children, newChildren)
	}

	if childrenOffer.Err() != nil {
		return childrenOffer.Err()
	}

	increasePriceParent := n

	for increasePriceParent != nil {
		increasePriceParent.len += total
		increasePriceParent.sum += sum
		increasePriceParent = increasePriceParent.parent
	}

	childrenCategory, err := stmtChildren.Query(n.id, categoryStr)
	if err != nil {
		return err
	}
	defer childrenCategory.Close()

	for childrenCategory.Next() {
		newChildren := &nodeItem{parent: n}

		var childrenId string
		if err = childrenCategory.Scan(&childrenId); err != nil {
			return err
		}

		if err = stmtItem.QueryRow(childrenId).Scan(&newChildren.id, &newChildren.parentId,
			&newChildren.name, &newChildren.price, &newChildren.Type, &newChildren.date); err != nil {
			return err
		}

		if err = newChildren.fill(stmtChildren, stmtItem); err != nil {
			return err
		}
		n.children = append(n.children, newChildren)
	}

	if childrenCategory.Err() != nil {
		return childrenCategory.Err()
	}

	return nil
}
