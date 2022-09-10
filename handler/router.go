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
		c.JSON(http.StatusBadRequest,
			gin.H{"code": http.StatusBadRequest, "message": "Validation failed"})
	} else if err = data.Import(json); err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": http.StatusInternalServerError, "message": "Internal server error"})
	} else {
		c.Status(http.StatusOK)
	}
}

func deleteRequest(c *gin.Context) {
	var uri schemas.DeleteRequest
	if err := c.ShouldBindUri(&uri); err != nil {
		c.Status(http.StatusConflict)
	}
	if err := data.Delete(uri.Id); err != nil {
		log.Println(err)
		c.Status(http.StatusConflict)
	}
}
