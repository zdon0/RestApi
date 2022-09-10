package handler

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func imports(c *gin.Context) {
	var json ImportRequest

	if err := c.ShouldBindJSON(&json); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "Validation failed"})
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
