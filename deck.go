package domino

import (
	"fmt"
	"math/rand"
	"time"
)

type Deck struct {
	offset int
	cards  [21]*Card
	last   int
	game   *Game
}

func NewDeck(game *Game) *Deck {
	return &Deck{game: game}
}

func (d *Deck) Shuffle() {
	index := 0
	for i := 1; i <= 6; i++ {
		for j := i; j <= 6; j++ {
			card := NewCard(i, j)
			card.SetBorder(true).
				SetTitle(fmt.Sprintf("[%d,%d]", i, j)).
				SetRect(0, 0, 10, 10)

			d.cards[index] = card
			index++
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})

	d.last = index
}

// print last n card in decks
func (d *Deck) PrintLastCards(num int) {
	size := len(d.cards)
	for i := 0; i < num; i++ {
		fmt.Println(d.cards[size-i-1])
	}
}

// pop n cards from deck
func (d *Deck) PopCards(num int) []*Card {
	if d.last-num < 0 {
		return nil
	}

	start := d.last - num
	end := start + num
	d.last = d.last - num

	newcards := make([]*Card, 0, num)
	newcards = append(newcards, d.cards[start:end]...)
	return newcards
}

// return number of cards in deck
func (d *Deck) GetNum() int {
	return d.last
}
