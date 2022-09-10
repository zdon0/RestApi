package handler

import (
	"RestApi/data"
	"RestApi/schemas"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func StartServer() {
	gin.DisableConsoleColor()
	InitValidators()

	r := gin.Default()
	r.POST("/imports", importRequest)
	r.DELETE("/delete/:id", deleteRequest)
	err := r.Run(":8080")
	log.Fatal(err.Error())
}

func importRequest(c *gin.Context) {
	json := &schemas.ImportRequest{}

	if err := c.ShouldBindJSON(json); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, schemas.BadRequest)
	} else if err = data.Import(json); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

func deleteRequest(c *gin.Context) {
	var uri schemas.DeleteRequest
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, schemas.BadRequest)
	}
	if err := data.Delete(uri.Id); err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, schemas.NotFound)
		} else {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
		}
	}
}
