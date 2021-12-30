package main

import "github.com/gusti-andika/domino"

func main() {

	game := domino.NewGame(domino.StandaloneMode)
	player1 := domino.NewPlayer("Player 1")
	player1.AssignCards(game.Deck.PopCards(5))
	game.Join(player1)

	player2 := domino.NewPlayer("Player 2")
	player2.AssignCards(game.Deck.PopCards(5))
	game.Join(player2)

	player3 := domino.NewPlayer("Player 3")
	player3.AssignCards(game.Deck.PopCards(5))
	game.Join(player3)
	game.Run()
}
