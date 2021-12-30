package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

type PlayerUI struct {
	*tview.Flex
	cards []*Card
	ID    string
}

func NewPlayerUI() *PlayerUI {
	return &PlayerUI{
		Flex: tview.NewFlex(),
	}
}

func (ui *PlayerUI) SetCards(cards []*Card) {
	ui.cards = cards
	ui.Clear()

	for _, card := range ui.cards {
		ui.AddItem(card, 10, 1, false)
	}
}

func (ui *PlayerUI) AddCard(card *Card) {
	ui.cards = append(ui.cards, card)
	ui.Clear()

	for _, card := range ui.cards {
		ui.AddItem(card, 0, 1, false)
	}
}

func (ui *PlayerUI) GetCardNum() int {
	return len(ui.cards)
}

func (ui *PlayerUI) GetCard(index int) *Card {
	return ui.cards[index]
}

func (ui *PlayerUI) PrintCards() {
	for _, c := range ui.cards {
		fmt.Printf("%v\n", c)
	}
}
