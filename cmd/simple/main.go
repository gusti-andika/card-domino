// Demo code which illustrates how to implement your own primitive.
package main

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/gusti-andika/domino"
	"github.com/rivo/tview"
)

func main() {
	deck := []tview.Primitive{}
	n := 10

	for n > 0 {
		x := rand.Intn(6) + 1
		y := rand.Intn(6) + 1
		card := domino.NewCard(x, y)

		deck = append(deck, card)
		n--
	}

	//grid := tview.NewGrid().SetBorders(true)
	app := tview.NewApplication()
	flex := tview.NewFlex()
	flex2 := tview.NewFlex().SetDirection(tview.FlexRow)
	for index, c := range deck {
		//fmt.Printf("%v\n", *c)
		row := index/3 + 1
		col := index%3 + 1

		c2, _ := c.(*domino.Card)
		fmt.Println("x: ", c2.X, "y:", c2.Y, " , ", row, " - ", col)
		//grid.AddItem(c, row, col, 1, 1, 0, 0, false)
		flex.AddItem(c, 0, 1, true)
	}
	flex.SetBorder(true).SetTitle("Played Cards")
	flex2.AddItem(flex, 0, 1, true)
	flex2.AddItem(tview.NewBox().SetBorder(true).SetTitle("Left (1/2 x width of Top)"), 0, 1, false)
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cycleFocus(app, deck, false)
		} else if event.Key() == tcell.KeyBacktab {
			cycleFocus(app, deck, true)
		}
		return event
	})
	if err := app.SetRoot(flex2, true).SetFocus(flex2).Run(); err != nil {
		panic(err)
	}
}

func cycleFocus(app *tview.Application, elements []tview.Primitive, reverse bool) {
	for i, el := range elements {
		if !el.HasFocus() {
			continue
		}

		if reverse {
			i = i - 1
			if i < 0 {
				i = len(elements) - 1
			}
		} else {
			i = i + 1
			i = i % len(elements)
		}

		app.SetFocus(elements[i])
		return
	}
}
