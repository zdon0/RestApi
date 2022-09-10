package handler

import (
	"RestApi/data"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"time"
)

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

func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validimport", validateParentsTypes)
	}
}

var validateParentsTypes validator.Func = func(fl validator.FieldLevel) bool {
	categories := map[string]bool{}
	offers := map[string]bool{}
	parents := map[string]bool{}

	for _, item := range fl.Field().Interface().([]ImportUnit) {
		if item.Type == "CATEGORY" {
			categories[item.Id] = true
		}
		if item.Type == "OFFER" {
			offers[item.Id] = true
		}
		parents[item.ParentId] = true
	}
	delete(parents, "")

	for key := range parents {
		if offers[key] {
			return false
		} else if categories[key] {
			delete(parents, key)
		}
	}
	return data.Validate(parents, offers, categories)
}
