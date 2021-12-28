package domino

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Game struct {
	*tview.Flex
	App     *tview.Application
	Players []*Player
	Deck    *Deck

	head           *Card
	tail           *Card
	last           *Card //last played card
	headView       *tview.Flex
	tailView       *tview.Flex
	size           int
	playedCardTail int
	playedCardHead int
	currentPlayer  int
	log            *LogWindow
	finish         bool
}

func NewGame() *Game {

	game := &Game{
		Flex:          tview.NewFlex().SetDirection(tview.FlexRow),
		headView:      tview.NewFlex(),
		tailView:      tview.NewFlex(),
		size:          0,
		currentPlayer: -1,
	}

	game.log = NewLogWindow(game)
	game.Players = make([]*Player, 3)
	game.Players[0] = NewPlayer(game, "Player 1")
	game.Players[1] = NewPlayer(game, "Player 2")
	game.Players[2] = NewPlayer(game, "Player 3")

	// setup UI & layout
	header := tview.NewFlex()
	header.AddItem(game.headView, 0, 1, false)
	game.headView.SetBorder(true).SetTitle("Head[First 3 Cards]")
	header.AddItem(game.tailView, 0, 1, false)
	game.tailView.SetBorder(true).SetTitle("Tail[Last 3 Cards]")
	game.AddItem(header, 0, 1, false)

	game.AddItem(game.Players[0], 0, 1, true)
	game.AddItem(game.Players[1], 0, 1, false)
	game.AddItem(game.Players[2], 0, 1, false)
	header.AddItem(game.log, 0, 1, false)

	// init deck and suffle cards
	game.Deck = NewDeck(game)
	game.Deck.Shuffle()

	// assign card to players
	game.Players[0].AssignCards(game.Deck.PopCards(5))
	game.Players[1].AssignCards(game.Deck.PopCards(5))
	game.Players[2].AssignCards(game.Deck.PopCards(5))

	// set current player
	game.currentPlayer = 0

	game.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if game.currentPlayer < 0 || game.finish {
			return event
		}

		els := make([]interface{}, len(game.Players[game.currentPlayer].cards))
		for i := range game.Players[game.currentPlayer].cards {
			els[i] = game.Players[game.currentPlayer].cards[i]
		}

		switch event.Key() {
		case tcell.KeyRight:
			selectCard(game, els, false)
		case tcell.KeyLeft:
			selectCard(game, els, true)
		case tcell.KeyEnter:
			game.update()
		}

		return event
	})

	return game
}

func (g *Game) playCard(card *Card) bool {
	isAppend := false
	switch {
	case g.size == 0:
		g.playedCardHead = card.X
		g.playedCardTail = card.Y

	case g.size >= 1:
		if g.playedCardHead == card.X {
			g.playedCardHead = card.Y
			card.Flip()
		} else if g.playedCardHead == card.Y {
			g.playedCardHead = card.X
		} else if g.playedCardTail == card.X {
			g.playedCardTail = card.Y
			isAppend = true
		} else if g.playedCardTail == card.Y {
			g.playedCardTail = card.X
			isAppend = true
			card.Flip()
		}

	}

	if g.head == nil {
		g.head, g.tail = card, card
	} else {
		if isAppend {
			card.prev = g.tail
			g.tail.next = card
			g.tail = card
		} else {
			old := g.head
			old.prev = card
			g.head = card
			g.head.next = old
		}
	}

	g.size++
	g.headView.Clear()
	next := g.head
	for i := 0; next != nil && i < 3; i++ {
		g.headView.AddItem(next, 10, 1, false)
		next = next.next
	}

	g.tailView.Clear()
	prev := g.tail
	for i := 0; prev.prev != nil && i < 2; i++ {
		prev = prev.prev
	}

	for prev != nil {
		g.tailView.AddItem(prev, 10, 1, false)
		prev = prev.next
	}

	if g.last != nil {
		g.last.ClearHighlight()
	}

	g.last = card
	g.last.Highlight()

	g.Log(fmt.Sprintf("head:%d tail: %d", g.playedCardHead, g.playedCardTail))

	return g.isFinish()
}

func (g *Game) validCard(card *Card) bool {

	switch {
	case card.Played:
		return false
	case g.size == 0:
		return true
	case g.size >= 1:
		if g.playedCardHead == card.X {
			return true
		} else if g.playedCardHead == card.Y {
			return true
		} else if g.playedCardTail == card.X {
			return true
		} else if g.playedCardTail == card.Y {
			return true
		} else {
			return false
		}
	}

	return false
}

func (g *Game) Run() {
	g.App = tview.NewApplication()

	g.log.SetDynamicColors(true)
	if err := g.App.SetRoot(g, true).SetFocus(g.CurrentPlayer()).Run(); err != nil {
		panic(err)
	}
}

func (g *Game) Log(s string) {
	s = fmt.Sprintf("[violet][sys[]:%s\n[white]", s)
	g.log.Write([]byte(s))
}

func (g *Game) CurrentPlayer() *Player {
	if g.currentPlayer < 0 {
		return nil
	}

	return g.Players[g.currentPlayer]
}

func (g *Game) SelectedCard() *Card {
	if g.currentPlayer < 0 || g.CurrentPlayer().selectedCard < 0 {
		return nil
	}

	return g.CurrentPlayer().cards[g.CurrentPlayer().selectedCard]
}

func (g *Game) update() {
	// check game is started when there's player playing
	if g.CurrentPlayer() == nil {
		return
	}

	if g.CurrentPlayer().HasPlayableCards() == false {
		g.Log(fmt.Sprintf("%s not have playable card. Skipping turn...", g.CurrentPlayer().id))
		g.nextPlayer()
		return
	}

	// check selected card is valid card
	if g.validCard(g.SelectedCard()) == false {
		if g.SelectedCard().Played {
			g.CurrentPlayer().Log(fmt.Sprintf("Card already played. Please select another "))
		} else {
			g.CurrentPlayer().Log(fmt.Sprintf("Card [%d - %d] not playable. Please select another ", g.SelectedCard().X, g.SelectedCard().Y))
		}

		return
	}

	// get current player played card and clone to playedcards array
	playedCard := g.CurrentPlayer().PlayCard()
	if g.playCard(NewCard(playedCard.X, playedCard.Y)) {
		g.end()
	} else {
		g.nextPlayer()
	}
}

func (g *Game) nextPlayer() {

	g.CurrentPlayer().selectedCard = -1
	g.CurrentPlayer().SetBorderColor(tcell.ColorWhite)
	g.currentPlayer++
	if g.currentPlayer >= len(g.Players) {
		g.currentPlayer = 0
	}

	g.App.SetFocus(g.CurrentPlayer())
	g.CurrentPlayer().SetBorderColor(tcell.ColorBlue)

	if !g.CurrentPlayer().HasPlayableCards() {
		g.Log(fmt.Sprintf("%s not have playable card. Skipping turn...", g.CurrentPlayer().id))
		g.nextPlayer()
	}
}

func (g *Game) isFinish() bool {
	finish := false

	// game is finished when either
	// 1. a player has played all his cards
	for _, player := range g.Players {
		if player.RemainingCardCount() == 0 {
			finish = true
			break
		}
	}

	// or
	// 2 all player not has any playable card
	if !finish {
		finish = true
		for _, player := range g.Players {
			if player.HasPlayableCards() {
				finish = false
				break
			}
		}
	}

	return finish
}

func (g *Game) end() {
	min := 1000
	var winner *Player
	for _, k := range g.Players {
		if k.RemainingCardValue() < min {
			min = k.RemainingCardValue()
			winner = k
		}
	}
	g.finish = true
	g.Log(fmt.Sprintf("[::bl]GAME FINISHED. Winner is [%s]%s", winner.color, winner.name))

}

func selectCard(g *Game, elements []interface{}, reverse bool) {
	var focusIdx int = 0
	for i, el2 := range elements {
		var card *Card = nil
		switch el2.(type) {
		case *Card:
			card = el2.(*Card)
		}
		if card != nil && !card.HasFocus() {
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

		focusIdx = i
	}

	if focusIdx < len(elements) {
		el3 := elements[focusIdx].(tview.Primitive)
		g.App.SetFocus(el3)
		g.CurrentPlayer().selectedCard = focusIdx
	}
}
