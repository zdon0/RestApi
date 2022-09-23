package main

import (
	"RestApi/data"
	"RestApi/router"
	"os"
)

func main() {
	var user, password string

	user = os.Getenv("PG_USER")
	password = os.Getenv("PG_PASSWORD")

	data.StartPG(user, password)
	router.StartServer()
}
