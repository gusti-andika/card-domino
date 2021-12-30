package domino

import (
	"fmt"

	"github.com/gusti-andika/domino/eventbus"
	"github.com/gusti-andika/domino/ui"
	"github.com/rivo/tview"
)

type GameMode int32

const (
	StandaloneMode GameMode = iota
	ServerMode
	ClientMode
)

type Game struct {
	App     *tview.Application
	Players []*Player
	Deck    *Deck

	ui            *ui.GameUI
	currentPlayer int
	finish        bool
	sendLogFunc   func(channel interface{}, msg string)
	//mode          GameMode
}

func NewGame(mode GameMode) *Game {
	game := &Game{
		currentPlayer: -1,
		//mode:          mode,
		ui:            ui.NewGameUI(),
	}

	// init deck and suffle cards
	game.Deck = NewDeck(game)
	game.Deck.Shuffle()

	game.App = tview.NewApplication()

	eventbus.SetLog(game.ui.GetLog())
	return game
}

func (g *Game) updateStatusView() {
	if g.CurrentPlayer() == nil {
		g.ui.ClearStatus()
		return
	}

	currentPlayer := g.CurrentPlayer()
	status := fmt.Sprintf("[black::b][CURRENT_PLAYER:[%s]%s][black::b] [HEAD:%d] [TAIL:%d]",
		currentPlayer.color, currentPlayer.Name, g.ui.GetHeadVal(), g.ui.GetTailVal())
	g.ui.UpdateStatus(status)
}

func (g *Game) Join(player *Player) bool {
	g.Log(fmt.Sprintf("%s joined", player.Name))
	
	if len(g.Players) >= 3 {
		g.Log(fmt.Sprintf("Can't join player: %s to game. Players already full ", player.Name))
		return false
	}

	player.game = g
	g.Players = append(g.Players, player)
	//g.Log(fmt.Sprintf("%s joined", player.Name))
	g.ui.AddPlayerUI(player.ui)

	eventData := map[string]interface{}{"player": player}
	eventbus.Post(eventbus.Event{eventbus.PlayerJoined, eventData})

	return true
}

func (g *Game) start() {
	var firstCard []*ui.Card
	for firstCard = g.Deck.PopCards(1); firstCard != nil; firstCard = g.Deck.PopCards(1) {
		playable := 0

		for _, p := range g.Players {
			b := p.IsPlayableFor(firstCard[0])
			if b {
				playable++
			}
		}

		if playable == len(g.Players) {
			g.ui.AddCard(firstCard[0])
			g.Log(fmt.Sprintf("Game Initiated with card [%d,%d]", firstCard[0].X, firstCard[0].Y))

			eventData := map[string]interface{}{"card": firstCard[0]}
			eventbus.Post(eventbus.Event{eventbus.InitialCard, eventData})
			return
		}
	}

	g.finish = true
	g.Log("Can not start game. Could not initiate playable card")
}

func (g *Game) playCard() (isFinish bool, playedCard *ui.Card, isAppend bool) {
	isAppend, isFinish = false, false
	playedCard = g.CurrentPlayer().PlayCard()
	if playedCard == nil {
		return
	}

	_, _, isAppend = g.ui.AddCard(ui.NewCard(playedCard.X, playedCard.Y))
	isFinish = g.isFinish()
	return
}

func (g *Game) Run() {
	InitGameEventHandlers(g)
	g.Log("Waiting for players...")
	
	go StartServer(g)

	func() {
		if err := g.App.SetRoot(g.ui, true).Run(); err != nil {
			panic(err)
		}
	}()
}

func (g *Game) GetPlayerById(id string) *Player {
	for _, p := range g.Players {
		if id == p.Id {
			return p
		}
	}

	return nil
}

func (g *Game) Log(s string) {
	s = fmt.Sprintf("[violet::r][sys[]:%s\n[white::-]", s)
	//data := map[string]interface{}{"msg": s}
	//eventbus.Post(eventbus.Event{eventbus.GameLog, data})
	 g.ui.Log(s)
	 //fmt.Print(s)
	// g.sendLog(s)
}

func (g *Game) sendLog(s string) {
	if g.sendLogFunc == nil {
		return
	}

	for _, p := range g.Players {
		if p.ClientChannel != nil {
			g.sendLogFunc(p.ClientChannel, s)
		}
	}
}

func (g *Game) CurrentPlayer() *Player {
	if g.currentPlayer < 0 {
		return nil
	}

	return g.Players[g.currentPlayer]
}

func (g *Game) SelectedCard() *ui.Card {
	if g.currentPlayer < 0 || g.CurrentPlayer().selectedCard < 0 {
		return nil
	}

	return g.CurrentPlayer().GetSelectedCard()
}

func (g *Game) update() {
	// check game is started when there's player playing
	if g.CurrentPlayer() == nil {
		return
	}

	// if !g.CurrentPlayer().HasPlayableCards() {
	// 	g.Log(fmt.Sprintf("%s not have playable card. Skipping turn...", g.CurrentPlayer().Id))
	// 	currentPlayer := g.CurrentPlayer()
	// 	nextPlayer := g.nextPlayer()
	// 	eventData := map[string]interface{}{"currentPlayer": currentPlayer, "nextPlayer": nextPlayer}
	// 	eventbus.Post(eventbus.Event{eventbus.SkipTurn, eventData})
	// 	return
	// }

	// check selected card is valid card
	// if !g.validCard(g.SelectedCard()) {
	// 	var msg string
	// 	if g.SelectedCard().Played {
	// 		msg = "Card already played. Please select another "
	// 		g.CurrentPlayer().Log(msg)
	// 	} else {
	// 		msg = fmt.Sprintf("Card [%d,%d] not playable. Please select another ", g.SelectedCard().X, g.SelectedCard().Y)
	// 		g.CurrentPlayer().Log(msg)
	// 	}

	// 	eventData := map[string]interface{}{"currentPlayer": g.CurrentPlayer(), "msg": msg}
	// 	eventbus.Post(eventbus.Event{eventbus.InvalidMove, eventData})
	// 	return
	// }

	// play current player selected card
	isGameFinish, playedCard, _ := g.playCard()
	// get current player
	currentPlayer := g.CurrentPlayer()
	// move to next player
	nextPlayer := g.nextPlayer()

	// notify player moved event
	eventData := map[string]interface{}{"currentPlayer": currentPlayer, "card": playedCard, "nextPlayer": nextPlayer}
	eventbus.Post(eventbus.Event{eventbus.PlayerMoved, eventData})

	if isGameFinish {
		eventData := map[string]interface{}{"winner": g.end()}
		eventbus.Post(eventbus.Event{eventbus.GameFinished, eventData})
	}

	g.updateStatusView()
}

func (g *Game) nextPlayer() *Player {

	if g.CurrentPlayer() != nil {
		g.CurrentPlayer().SetCurrentPlayer(false)
	}

	g.currentPlayer++
	if g.currentPlayer >= len(g.Players) {
		g.currentPlayer = 0
	}

	g.CurrentPlayer().SetCurrentPlayer(true)

	if !g.CurrentPlayer().HasPlayableCards() {
		g.Log(fmt.Sprintf("%s not have playable card. Skipping turn...", g.CurrentPlayer().Id))
		eventData := map[string]interface{}{"currentPlayer": g.CurrentPlayer()}
		eventbus.Post(eventbus.Event{eventbus.SkipTurn, eventData})

		return g.nextPlayer()
	}

	return g.CurrentPlayer()
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

func (g *Game) end() *Player {
	min := 1000
	var winner *Player
	for _, k := range g.Players {
		if k.RemainingCardValue() < min {
			min = k.RemainingCardValue()
			winner = k
		}
	}
	g.finish = true
	g.Log(fmt.Sprintf("[::bl]GAME FINISHED. Winner is [%s]%s", winner.color, winner.Name))
	return winner
}
