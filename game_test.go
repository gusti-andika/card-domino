package domino

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGameInit(t *testing.T) {
	game := NewGame()
	game.Join("player1")
	game.Join("player2")
	comparer := cmp.Comparer(func(a, b *Card) bool {
		return a.X == b.X && a.Y == b.Y
	})

	if diff := cmp.Diff(game.Players[0].cards, game.Players[1].cards, comparer); diff == "" {
		t.Error("Player1 and player2 should not have same cards")
	}
}

func TestValidCard(t *testing.T) {
	game := NewGame()
	card := NewCard(1, 1)

	// test first played card is always valid
	if game.size == 0 && !game.validCard(card) {
		t.Fatalf("Expecting card : %+v to be valid but not", card)
	}

	// test second played card valid if only matched previous played card
	// assumed first  played card is [6 - 6], then [1 - 1] is invalid but [6-?] or [?-6] is valid ones

	game.playCard(NewCard(6, 6))

	if game.validCard(NewCard(1, 2)) != false {
		t.Fatalf("Expecting [1,2] to be invalid")
	}

	if game.validCard(NewCard(6, 2)) != true {
		t.Fatalf("Expecting [6,2]to be valid")
	}

	if game.validCard(NewCard(2, 6)) != true {
		t.Fatalf("Expecting [2,6] to be valid")
	}

	// test third played card valid if only matched previous played card
	// after [6-6] then [2-6] is played next valid one is [6-?] or [?-6] or [2-?] or [?-2]
	game.playCard(NewCard(2, 6))

	if game.validCard(NewCard(6, 3)) != true {
		t.Fatalf("Expecting [6,3] to be valid")
	}

	if game.validCard(NewCard(4, 6)) != true {
		t.Fatalf("Expecting [4,6]to be valid")
	}

	if game.validCard(NewCard(2, 6)) != true {
		t.Fatalf("Expecting [2,6] to be valid")
	}

	if game.validCard(NewCard(2, 2)) != true {
		t.Fatalf("Expecting [2,2] to be valid")
	}

	if game.validCard(NewCard(3, 3)) != false {
		t.Fatalf("Expecting [3,3] to be invalid")
	}
}
