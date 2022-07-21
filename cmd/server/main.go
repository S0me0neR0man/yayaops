package main

import (
	"github.com/S0me0neR0man/yayaops/internal/server"
	"log"
)

func main() {
	err := server.New().Start()
	if err != nil {
		log.Println(err)
	}
}
