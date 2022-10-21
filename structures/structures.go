package structures

import (
	"github.com/gofrs/uuid"
	"time"
)

type ImportUnit struct {
	Id       uuid.NullUUID `json:"id" binding:"required,uuid_rfc4122"`
	Name     string        `json:"name" binding:"required"`
	ParentId uuid.NullUUID `json:"parentId" binding:"omitempty,uuid_rfc4122"`
	Type     string        `json:"type" binding:"required,oneof=OFFER CATEGORY"`
	Price    int           `json:"price" binding:"excluded_unless=Type CATEGORY,required_if=Type OFFER,gte=0"`
}

type ImportRequest struct {
	Items      []ImportUnit `json:"items" binding:"required,unique=Id,validImport,dive,required"`
	UpdateDate time.Time    `json:"updateDate" time_format:"2006-01-02T03:04:05.000Z" binding:"required"`
}

type SalesRequest struct {
	Date time.Time `json:"date" form:"date" time_format:"2006-01-02T03:04:05.000Z" binding:"required,validDate"`
}

type IdRequest struct {
	Id string `uri:"id" binding:"required,uuid_rfc4122"`
}

var NotFound = map[string]any{"code": 404, "message": "Item not found"}
var BadRequest = map[string]any{"code": 400, "message": "Validation failed"}
