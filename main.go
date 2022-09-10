package main

import (
	"RestApi/data"
	"RestApi/handler"
)

func main() {
	data.StartPG("postgres", "postgres")
	handler.StartServer()
}
