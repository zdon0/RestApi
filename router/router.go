package router

import (
	"RestApi/data"
	"RestApi/schemas"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func StartServer() {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	InitValidators()

	r := gin.Default()
	r.TrustedPlatform = gin.PlatformGoogleAppEngine
	r.POST("/imports", importRequest)
	r.DELETE("/delete/:id", deleteRequest)
	r.GET("/nodes/:id", nodesRequest)
	r.GET("/sales", salesRequest)
	if err := endless.ListenAndServe(":8080", r); err != nil {
		log.Println(err)
	}
	log.Println("server stopped")
}

func importRequest(c *gin.Context) {
	json := &schemas.ImportRequest{}

	if err := c.ShouldBindJSON(json); err != nil {

		c.JSON(http.StatusBadRequest, schemas.BadRequest)

	} else if err = data.Import(json); err != nil {

		log.Println(err)
		c.Status(http.StatusInternalServerError)

	}
}

func deleteRequest(c *gin.Context) {
	var uri schemas.IdRequest
	if err := c.ShouldBindUri(&uri); err != nil {

		c.JSON(http.StatusBadRequest, schemas.BadRequest)

	} else if err = data.Delete(uri.Id); err != nil {
		if err.Error() == "not found" {

			c.JSON(http.StatusNotFound, schemas.NotFound)

		} else {

			log.Println(err)
			c.Status(http.StatusInternalServerError)

		}
	}
}

func nodesRequest(c *gin.Context) {
	var uri schemas.IdRequest

	if err := c.ShouldBindUri(&uri); err != nil {

		c.JSON(http.StatusBadRequest, schemas.BadRequest)

	} else if res, err := data.Nodes(uri.Id); err != nil {

		if err.Error() == "not found" {

			c.JSON(http.StatusNotFound, schemas.NotFound)

		} else {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
		}

	} else {
		c.JSON(http.StatusOK, res)
	}
}

func salesRequest(c *gin.Context) {
	var date schemas.SalesRequest

	if err := c.ShouldBind(&date); err != nil {
		c.JSON(http.StatusBadRequest, schemas.BadRequest)
	} else if res, err := data.Sales(date.Date); err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, res)
	}
}
