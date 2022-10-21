package main

import (
	"RestApi/data"
	"RestApi/router"
)

func main() {
	data.StartPG()
	router.StartServer()
}
