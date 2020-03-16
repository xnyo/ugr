package main

import (
	"log"
	"os"

	"github.com/xnyo/ugr/bot"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide the token")
	}
	err := bot.Initialize(os.Args[1])
	if err != nil {
		log.Fatal(err)
		return
	}
	bot.Start()
}
