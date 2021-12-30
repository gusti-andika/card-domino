package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Card implements a simple primitive for radio button selections.
type Card struct {
	*tview.Box
	X                 int
	Y                 int
	Played            bool
	flipped           bool
	higlighted        bool
	next, prev        *Card
	hideNotPlayedCard bool
}

// NewCard returns a new radio button primitive.
func NewCard(x int, y int) *Card {
	card := &Card{
		Box: tview.NewBox(),
		X:   x,
		Y:   y,
	}

	card.SetBorder(true).SetBorderColor(tcell.ColorBlue).
		SetTitle(fmt.Sprintf("[%d,%d]", x, y)).
		SetRect(0, 0, 10, 10)
	return card
}

func (card *Card) Highlight() {
	card.SetBorderColor(tcell.ColorYellow)
	card.higlighted = true
}

func (card *Card) ClearHighlight() {
	if card.Played {
		card.SetBorderColor(tcell.ColorRed)
	} else {
		card.SetBorderColor(tcell.ColorBlue)
	}

	if !card.Played && card.hideNotPlayedCard {
		card.SetTitle("[?,?]")
	} else {
		card.SetTitle(fmt.Sprintf("[%d,%d]", card.X, card.Y))
	}
	card.higlighted = false
}

func (card *Card) Flip() {
	if card.flipped {
		return
	}

	card.X, card.Y = card.Y, card.X
	card.SetTitle(fmt.Sprintf("[%d,%d]", card.X, card.Y))
	card.flipped = true
}

func (card *Card) Play() {
	card.Played = true
	card.ClearHighlight()
}

func (card *Card) SetValue(x int, y int) {
	card.X, card.Y = x, y
	card.ClearHighlight()
}

func (card *Card) SetHideNotPlayedCards(hide bool) {
	card.hideNotPlayedCard = hide
	card.ClearHighlight()
}

// Draw draws this primitive onto the screen.
func (r *Card) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()
	curY, curX, check, i := y, x, rune('\u25c9'), 0
	offsetX, offsetY := ((width/2)/2)+x, 0

	if !r.Played && r.hideNotPlayedCard {
		return
	}

	for i < r.X {
		curY = i / 2
		curX = i % 2
		screen.SetContent(curX*(width/2)+offsetX, offsetY+y+curY, check, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
		i++
	}

	// Draw a horizontal line across the middle of the box.
	offsetY = height / 2
	for cx := x + 1; cx < x+width-1; cx++ {
		screen.SetContent(cx, offsetY+y-1, tview.BoxDrawingsHeavyHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	}

	i = 0
	for i < r.Y {
		curY = i / 2
		curX = i % 2
		screen.SetContent(curX*(width/2)+offsetX, offsetY+y+curY, check, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
		i++
	}
}
