package main

import "github.com/gusti-andika/domino"

func main() {

	game := domino.NewGame()
	game.Join("Player 1")
	game.Join("Player 2")
	game.Join("Player 3")
	game.Run()
}
