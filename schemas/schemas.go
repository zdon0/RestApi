package schemas

import "time"

type ImportUnit struct {
	Id       string `json:"id" binding:"required,uuid_rfc4122"`
	Name     string `json:"name" binding:"required"`
	ParentId string `json:"parentId" binding:"omitempty,uuid_rfc4122"`
	Type     string `json:"type" binding:"required,oneof=OFFER CATEGORY"`
	Price    int    `json:"price" binding:"excluded_unless=Type CATEGORY,required_if=Type OFFER,gte=0"`
}

type ImportRequest struct {
	Items      []ImportUnit `json:"items" binding:"required,unique=Id,validimport,dive,required"`
	UpdateDate time.Time    `json:"updateDate" time_format:"2006-01-02T03:04:05.000Z" binding:"required"`
}
