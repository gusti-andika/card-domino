package main

import "github.com/gusti-andika/domino"

func main() {

	game := domino.NewGame()
	game.Join("Player 1", false)
	game.Join("Player 2", true)
	game.Join("Player 3", true)
	game.Run()
}
