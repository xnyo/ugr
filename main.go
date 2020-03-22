package main

import (
	"log"

	"github.com/xnyo/ugr/bot"
)

func main() {
	err := bot.Initialize()
	if err != nil {
		log.Fatal(err)
		return
	}
	bot.Start()
}
