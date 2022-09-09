package main

import (
	"RestApi/data"
	"log"
)

func main() {
	if err := data.Start("postgres", "postgres"); err != nil {
		log.Fatal(err.Error())
	}
}
