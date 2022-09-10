package handler

import (
	"RestApi/data"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"time"
)

type Import struct {
	Id       string `json:"id" binding:"required,uuid_rfc4122"`
	Name     string `json:"name" binding:"required"`
	ParentId string `json:"parentId" binding:"omitempty,uuid_rfc4122"`
	Type     string `json:"type" binding:"required,oneof=OFFER CATEGORY"`
	Price    int    `json:"price" binding:"excluded_unless=Type CATEGORY,required_if=Type OFFER,gte=0"`
}

type ImportRequest struct {
	Items      []Import  `json:"items" binding:"required,unique=Id,validparents,dive,required"`
	UpdateDate time.Time `json:"updateDate" time_format:"2006-01-02T03:04:05.000Z" binding:"required"`
}

func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validparents", validateParents)
	}
}

var validateParents validator.Func = func(fl validator.FieldLevel) bool {
	categories := map[string]bool{}
	offers := map[string]bool{}
	check := map[string]bool{}

	for _, item := range fl.Field().Interface().([]Import) {
		if item.Type == "CATEGORY" {
			categories[item.Id] = true
		}
		if item.Type == "OFFER" && item.ParentId != "" {
			offers[item.Id] = true
		}
		check[item.ParentId] = true
	}
	delete(check, "")

	for key, _ := range check {
		if offers[key] {
			return false
		} else if categories[key] {
			delete(check, key)
		}
	}
	if len(check) == 0 {
		return true
	}
	return data.AreParents(check)
}
