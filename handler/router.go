package handler

import (
	"RestApi/data"
	"RestApi/schemas"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func imports(c *gin.Context) {
	var json schemas.ImportRequest

	if err := c.ShouldBindJSON(&json); err != nil {
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

func StartServer() {
	gin.DisableConsoleColor()
	InitValidators()

	r := gin.Default()
	r.GET("/imports", imports)
	err := r.Run(":8080")
	log.Fatal(err.Error())
}
