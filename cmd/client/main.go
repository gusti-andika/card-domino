package main

import (
	"flag"
	"log"
	"strings"

	"github.com/gusti-andika/domino"
)

var playerName = flag.String("playerName", "", "player name")

func main() {
	flag.Parse()
	if len(strings.TrimSpace(*playerName)) == 0 {
		log.Printf("playerName can not empty")
		flag.Usage()
		return
	}

	client := domino.NewClient(*playerName)
	client.Run()
}
