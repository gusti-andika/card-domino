package main

import "github.com/gusti-andika/domino"

func main() {
	game := domino.NewGame(domino.ServerMode)
	game.Run()
}
