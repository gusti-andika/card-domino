package domino

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gusti-andika/domino/ui"
)

func TestShuffleAndPopCards(t *testing.T) {
	d := NewDeck(&Game{})
	d.Shuffle()

	if d.GetNum() != 21 {
		t.Errorf("Deck expecting to have %d cards but got %d", 36, d.GetNum())
	}

	//d.PrintLastCards(5)
	offset := d.GetNum() - 5
	expectedLast5cards := d.cards[offset : offset+5]
	last5cards := d.PopCards(5)

	if d.GetNum() != 16 {
		t.Errorf("After popped last 5 cards expecting to have 16 cards in deck but got %d cards", d.GetNum())
	}

	comparer := cmp.Comparer(func(a, b *ui.Card) bool {
		return a.X == b.X && a.Y == b.Y
	})
	if diff := cmp.Diff(expectedLast5cards, last5cards, comparer); diff != "" {
		t.Errorf("Wrong result on pop last 5 cards after shuffle")
	}

	if len(last5cards) != 5 {
		t.Errorf("Expected Deck.PopCards(5) only return 5 cards but got %d", len(last5cards))
	}

	last5cards = d.PopCards(5)
	if len(last5cards) != 5 {
		t.Errorf("Expected 2nd Deck.PopCards(5) only return 5 cards but got %d", len(last5cards))
	}

	if d.GetNum() != 11 {
		t.Errorf("After popped last 5 cards expecting to have 11 cards in deck but got %d cards", d.GetNum())
	}

}
