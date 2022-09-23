package main

import (
	"RestApi/data"
	"RestApi/router"
	"os"
)

func main() {
	var port, user, password string
	var exist bool

	if port, exist = os.LookupEnv("PORT"); !exist {
		port = "8080"
	}

	if user, exist = os.LookupEnv("PG_USER"); !exist {
		user = "postgres"
	}

	if password, exist = os.LookupEnv("PG_PASSWORD"); !exist {
		password = "postgres"
	}

	data.StartPG(user, password)
	router.StartServer(port)
}
