package main

import (
	"RestApi/data"
	"RestApi/handler"
	"flag"
)

func main() {
	port := flag.String("port", "8080", "set server port")
	login := flag.String("login", "postgres", "database login")
	password := flag.String("password", "postgres", "database password")
	flag.Parse()

	data.StartPG(*login, *password)
	handler.StartServer(*port)
}
