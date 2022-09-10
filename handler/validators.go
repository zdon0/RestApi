package handler

import (
	"RestApi/data"
	"RestApi/schemas"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validimport", validateParentsTypes)
	}
}

var validateParentsTypes validator.Func = func(fl validator.FieldLevel) bool {
	categories := map[string]bool{}
	offers := map[string]bool{}
	parents := map[string]bool{}

	for _, item := range fl.Field().Interface().([]schemas.ImportUnit) {
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
	return data.ValidateImport(parents, offers, categories)
}
