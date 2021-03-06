package domino

import (
	"fmt"
	"time"

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
	statusView     *tview.TextView
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

	// setup UI & layout
	header := tview.NewFlex()
	header.AddItem(game.headView, 0, 1, false)
	game.headView.SetBorder(true).SetTitle("Head[First 3 Cards]")
	header.AddItem(game.tailView, 0, 1, false)
	game.tailView.SetBorder(true).SetTitle("Tail[Last 3 Cards]")
	game.AddItem(header, 0, 1, false)

	game.statusView = tview.NewTextView().SetDynamicColors(true)
	game.statusView.SetBackgroundColor(tcell.ColorYellow)
	logPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	logPanel.SetBorder(true).SetTitle("Log")
	logPanel.AddItem(game.log, 0, 1, false)
	logPanel.AddItem(game.statusView, 1, 1, false)
	header.AddItem(logPanel, 0, 1, false)

	// init deck and suffle cards
	game.Deck = NewDeck(game)
	game.Deck.Shuffle()

	game.Log("Waiting for players...")
	game.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if game.CurrentPlayer() == nil || game.finish {
			return event
		}

		player := game.CurrentPlayer()
		if player.isCpu {
			return event
		}

		switch event.Key() {
		case tcell.KeyRight:
			player.selectCard(false)
		case tcell.KeyLeft:
			player.selectCard(true)
		case tcell.KeyEnter:
			game.update()
		}

		return event
	})
	return game
}

func (g *Game) updateStatusView() {
	if g.CurrentPlayer() == nil {
		g.statusView.Clear()
		return
	}

	g.statusView.SetText(fmt.Sprintf("[black::b][CURRENT_PLAYER:[%s]%s][black::b] [HEAD:%d] [TAIL:%d]", g.CurrentPlayer().color, g.CurrentPlayer().name, g.playedCardHead, g.playedCardTail))
}

func (g *Game) Join(playerName string, isCpu bool) {
	if len(g.Players) >= 3 {
		g.Log(fmt.Sprintf("Can't join player: %s to game. Players already full ", playerName))
		return
	}

	player := NewPlayer(g, playerName, isCpu)
	player.isCpu = isCpu
	if player.isCpu {
		player.SetFocusFunc(func() {
			if !player.HasPlayableCards() {
				return
			}

			go func() {
				ch := make(chan struct{})
				for {
					select {
					case <-ch:
						goto finish
					default:
						g.App.QueueUpdateDraw(func() {

							selected := player.selectCard(false)
							time.Sleep(1000 * time.Millisecond)
							if g.validCard(selected) {
								g.update()
								close(ch)
							}
						})
					}
				finish:
				}
			}()

		})

		player.SetBlurFunc(func() {
			time.Sleep(1 * time.Second)
		})
	}

	player.AssignCards(g.Deck.PopCards(5))
	g.Players = append(g.Players, player)
	g.Log(fmt.Sprintf("%s joined", playerName))

	g.AddItem(player, 0, 1, false)

	// players acquired, start game
	if len(g.Players) == 3 {
		g.start()
	}
}

func (g *Game) start() {
	g.App = tview.NewApplication()
	var firstCard []*Card
	for firstCard = g.Deck.PopCards(1); firstCard != nil; firstCard = g.Deck.PopCards(1) {
		playable := 0

		for _, p := range g.Players {
			b := p.IsPlayableFor(firstCard[0])
			if b {
				playable++
			}

		}

		if playable == len(g.Players) {
			g.playCard(firstCard[0])
			g.nextPlayer()
			g.Log(fmt.Sprintf("Game Initiated with card [%d,%d]", firstCard[0].X, firstCard[0].Y))
			g.updateStatusView()
			return
		}
	}

	g.finish = true
	g.Log("Can not start game. Could not initiate playable card")
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
	g.log.SetDynamicColors(true)
	if err := g.App.SetRoot(g, true).Run(); err != nil {
		panic(err)
	}
}

func (g *Game) Log(s string) {
	s = fmt.Sprintf("[violet::r][sys[]:%s\n[white::-]", s)
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
	// if g.CurrentPlayer() == nil {
	// 	return
	// }

	// if !g.CurrentPlayer().HasPlayableCards() {
	// 	g.Log(fmt.Sprintf("%s not have playable card. Skipping turn...", g.CurrentPlayer().id))
	// 	g.nextPlayer()
	// 	return
	// }

	// check selected card is valid card
	if !g.validCard(g.SelectedCard()) {
		if g.SelectedCard().Played {
			g.CurrentPlayer().Log(fmt.Sprintf("Card [%d,%d] already played. Please select another ", g.SelectedCard().X, g.SelectedCard().Y))
		} else {
			g.CurrentPlayer().Log(fmt.Sprintf("Card [%d,%d] not playable. Please select another ", g.SelectedCard().X, g.SelectedCard().Y))
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

	g.updateStatusView()
}

func (g *Game) nextPlayer() {
	if g.CurrentPlayer() != nil {
		g.CurrentPlayer().selectedCard = -1
		g.CurrentPlayer().SetBorderColor(tcell.ColorWhite)
	}

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
