package domino

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gusti-andika/domino/eventbus"
	"github.com/gusti-andika/domino/ui"
)

var idCounter = 0
var lastColor = -1
var colors = [...]string{
	"maroon",
	"green",
	"olive",
	"navy",
	"purple",
	"teal",
	"silver",
	"gray",
	"red",
	"lime",
	"yellow",
	"blue",
	"fuchsia",
	"aqua",
	"white",
	"antiquewhite",
	"aquamarine",
	"azure",
	"beige",
	"bisque",
	"blanchedalmond",
	"blueviolet",
	"brown",
	"burlywood",
	"cadetblue",
	"chartreuse",
	"chocolate",
	"coral",
	"cornflowerblue",
	"cornsilk",
}

type Player struct {
	//ClientChannel rpc.GameService_UpdateServer
	Name          string
	Id            string
	ui            *ui.PlayerUI
	game          *Game
	selectedCard  int
	color         string
	remainingCard int
	allowInput    bool
	out           chan interface{}
	isMe          bool
}

func NewClientOpponent(name string, id string, color string) *Player {
	opponent := NewPlayer(name)
	opponent.isMe = false
	opponent.color = color
	opponent.ui.SetTitleColor(tcell.ColorNames[color])
	opponent.Id = id
	opponent.ui.SetBorder(true).SetTitle(fmt.Sprintf("%s[%s]", opponent.Name, opponent.Id))

	return opponent
}

func NewClientPlayer(name string) *Player {
	player := NewPlayer(name)
	player.ui.SetTitle(fmt.Sprintf("%s[%s]", player.Name, player.Id))
	return player
}

func NewPlayer(name string) *Player {
	player := &Player{
		ui:           ui.NewPlayerUI(),
		selectedCard: -1,
		Name:         name,
		out:          make(chan interface{}, 10),
		isMe:         true,
	}

	if lastColor == -1 {
		rand.Seed(time.Now().UnixNano())
		lastColor = rand.Intn(len(colors))
	} else {
		lastColor++
		if lastColor >= len(colors) {
			lastColor = 0
		}
	}
	player.color = colors[lastColor]
	player.ui.SetTitleColor(tcell.ColorNames[colors[lastColor]])
	idCounter++
	player.Id = fmt.Sprintf("P%d", idCounter)
	player.ui.SetBorder(true).SetTitle(fmt.Sprintf("You [%s]", player.Id))

	// player.ui.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	if !player.allowInput {
	// 		return event
	// 	}

	// 	switch event.Key() {
	// 	case tcell.KeyRight:
	// 		player.selectCard(false)
	// 	case tcell.KeyLeft:
	// 		player.selectCard(true)
	// 	case tcell.KeyEnter:
	// 		player.game.update()
	// 	}

	// 	return event
	// })

	return player
}

func (p *Player) SetId(id string) {
	p.Id = id
	name := p.Name
	if p.isMe {
		name = "You"
	}
	p.ui.SetTitle(fmt.Sprintf("%s [%s]", name, p.Id))

}

func (p *Player) SetColor(color string) {
	p.color = color
	p.ui.SetTitleColor(tcell.ColorNames[p.color])

}

func (p *Player) AllowInput(allowInput bool) {
	p.allowInput = allowInput
}

func (p *Player) AssignCards(cards []*ui.Card) {
	p.ui.SetCards(cards)

	p.game.Log("send card assigned event")
	eventData := map[string]interface{}{"cards": cards, "playerId": p.Id}
	eventbus.Post(eventbus.Event{Type: eventbus.CardAssigned, Data: eventData})
}

func (p *Player) SelectCard(card int) bool {
	if card >= p.ui.GetCardNum() {
		return false
	}

	p.selectedCard = card
	return true
}

func (p *Player) PrintCard() {
	p.ui.PrintCards()
}

func (p *Player) GetUI() *ui.PlayerUI {
	return p.ui
}

func (p *Player) GetSelectedCardIndex() int {
	return p.selectedCard
}

func (p *Player) GetSelectedCard() *ui.Card {
	if p.selectedCard < 0 || p.selectedCard >= p.ui.GetCardNum() {
		return nil
	}

	return p.ui.GetCard(p.selectedCard)
}

func (p *Player) PlayCard() *ui.Card {
	selectedCard := p.GetSelectedCard()
	if selectedCard == nil {
		return nil
	}

	selectedCard.Play()
	p.remainingCard--
	p.Log(fmt.Sprintf("Played card [%d,%d]", selectedCard.X, selectedCard.Y))

	return selectedCard
}

func (p *Player) Log(s string) {
	s = fmt.Sprintf("[%s][%s[]:%s\n[white]", p.color, p.Id, s)

	data := map[string]interface{}{"msg": s}
	eventbus.Post(eventbus.Event{eventbus.GameLog, data})
}

func (p *Player) HasPlayableCards() bool {
	valid := false
	for i := 0; i < p.ui.GetCardNum(); i++ {
		c := p.ui.GetCard(i)
		if c.Played {
			continue
		}

		if p.game.ui.ValidCard(c) {
			valid = true
			break
		}
	}

	return valid
}

func (p *Player) IsPlayableFor(card *ui.Card) bool {
	for i := 0; i < p.ui.GetCardNum(); i++ {
		c := p.ui.GetCard(i)
		switch {
		case card.Played:
			continue
		default:
			if c.X == card.X || c.X == card.Y {
				return true
			} else if c.Y == card.X || c.Y == card.Y {
				return true
			}
		}
	}

	return false
}

func (p *Player) RemainingCardValue() int {
	total := 0
	for i := 0; i < p.ui.GetCardNum(); i++ {
		c := p.ui.GetCard(i)
		if c.Played {
			continue
		}

		total += c.X + c.Y
	}
	return total
}

func (p *Player) RemainingCardCount() int {
	return p.remainingCard
}

func (p *Player) SetCurrentPlayer(current bool) {
	if current {
		p.game.App.SetFocus(p.ui)
		p.ui.SetBorderColor(tcell.ColorBlue)

	} else {
		p.selectedCard = -1
		p.ui.SetBorderColor(tcell.ColorWhite)
	}
}

func (p *Player) GetCardNum() int {
	return p.ui.GetCardNum()
}

func (p *Player) navigateCard(reverse bool) {
	var focusIdx int = 0
	for i := 0; i < p.ui.GetCardNum(); i++ {
		card := p.ui.GetCard(i)
		if !card.HasFocus() {
			continue
		}

		if reverse {
			i = i - 1
			if i < 0 {
				i = p.ui.GetCardNum() - 1
			}
		} else {
			i = i + 1
			i = i % p.ui.GetCardNum()
		}

		focusIdx = i
	}

	p.SelectCard(focusIdx)
}

func (p *Player) navigateRight() {
	if p.selectedCard >= p.ui.GetCardNum()-1 {
		return
	}
	//	var focusIdx int = 0
	p.selectedCard++

	//p.SelectCard(focusIdx)
}

func (p *Player) navigateLeft() {
	if p.selectedCard <= 0 {
		return
	}
	p.selectedCard--

}
