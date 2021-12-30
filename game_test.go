package domino

import (
	"testing"

	"github.com/gusti-andika/domino/ui"
)

func TestValidCard(t *testing.T) {
	game := NewGame(StandaloneMode)
	card := ui.NewCard(1, 1)

	// test first played card is always valid
	if game.ui.GetCardNum() == 0 && !game.ui.ValidCard(card) {
		t.Fatalf("Expecting card : %+v to be valid but not", card)
	}

	// test second played card valid if only matched previous played card
	// assumed first  played card is [6 - 6], then [1 - 1] is invalid but [6-?] or [?-6] is valid ones

	game.ui.AddCard(ui.NewCard(6, 6))

	if game.ui.ValidCard(ui.NewCard(1, 2)) != false {
		t.Fatalf("Expecting [1,2] to be invalid")
	}

	if game.ui.ValidCard(ui.NewCard(6, 2)) != true {
		t.Fatalf("Expecting [6,2]to be valid")
	}

	if game.ui.ValidCard(ui.NewCard(2, 6)) != true {
		t.Fatalf("Expecting [2,6] to be valid")
	}

	// test third played card valid if only matched previous played card
	// after [6-6] then [2-6] is played next valid one is [6-?] or [?-6] or [2-?] or [?-2]
	game.ui.AddCard(ui.NewCard(2, 6))

	if game.ui.ValidCard(ui.NewCard(6, 3)) != true {
		t.Fatalf("Expecting [6,3] to be valid")
	}

	if game.ui.ValidCard(ui.NewCard(4, 6)) != true {
		t.Fatalf("Expecting [4,6]to be valid")
	}

	if game.ui.ValidCard(ui.NewCard(2, 6)) != true {
		t.Fatalf("Expecting [2,6] to be valid")
	}

	if game.ui.ValidCard(ui.NewCard(2, 2)) != true {
		t.Fatalf("Expecting [2,2] to be valid")
	}

	if game.ui.ValidCard(ui.NewCard(3, 3)) != false {
		t.Fatalf("Expecting [3,3] to be invalid")
	}
}
