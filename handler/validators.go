package handler

import (
	"RestApi/data"
	"RestApi/schemas"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"time"
)

func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validImport", validateParentsTypes)
		v.RegisterValidation("validDate", validateDate)
	}
}

func validateParentsTypes(fl validator.FieldLevel) bool {
	categories := map[uuid.NullUUID]bool{}
	offers := map[uuid.NullUUID]bool{}
	parents := map[uuid.NullUUID]bool{}

	for _, item := range fl.Field().Interface().([]schemas.ImportUnit) {
		if item.Type == "CATEGORY" {
			categories[item.Id] = true
		}
		if item.Type == "OFFER" {
			offers[item.Id] = true
		}
		parents[item.ParentId] = true
	}
	delete(parents, uuid.NullUUID{Valid: false})

	for key := range parents {
		if offers[key] {
			return false
		} else if categories[key] {
			delete(parents, key)
		}
	}
	return data.ValidateImport(parents, offers, categories)
}

func validateDate(fl validator.FieldLevel) bool {
	date := fl.Field().Interface().(time.Time)
	return time.Now().After(date)
}
