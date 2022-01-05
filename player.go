package domino

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var idCounter, cpuCounter = 0, 0
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
	*tview.Flex
	cards         []*Card
	game          *Game
	selectedCard  int
	color         string
	name          string
	id            string
	remainingCard int
	isCpu         bool
}

func NewPlayer(game *Game, name string, isCpu bool) *Player {
	player := &Player{
		Flex:         tview.NewFlex(),
		game:         game,
		selectedCard: 0,
		name:         name,
		isCpu:        isCpu,
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
	player.SetTitleColor(tcell.ColorNames[colors[lastColor]])

	if !isCpu {
		idCounter++
		player.id = fmt.Sprintf("P%d", idCounter)

	} else {
		cpuCounter++
		player.id = fmt.Sprintf("CPU-%d", cpuCounter)

	}

	player.SetBorder(true).SetTitle(fmt.Sprintf("%s[%s]", player.name, player.id))
	return player
}

func (p *Player) AssignCards(cards []*Card) {
	p.cards = cards
	p.remainingCard = len(cards)
	if p.isCpu {
		for _, c := range p.cards {
			c.hideNotPlayedCard = true
			c.SetTitle("[?,?]")
		}
	}
	p.refresh()
}

func (p *Player) PrintCard() {
	for _, c := range p.cards {
		fmt.Printf("%v\n", c)
	}
}

func (p *Player) refresh() {
	p.Clear()
	for i := 0; i < p.remainingCard; i++ {
		p.AddItem(p.cards[i], 10, 1, false)
	}
}

func (p *Player) PlayCard() *Card {
	if p.selectedCard >= len(p.cards) {
		return nil
	}

	playedCard := p.cards[p.selectedCard]
	playedCard.Play()
	p.remainingCard--
	p.Log(fmt.Sprintf("Played card [%d,%d]", playedCard.X, playedCard.Y))
	return playedCard
}

func (p *Player) Log(s string) {
	s2 := fmt.Sprintf("[%s][%s[]:%s\n[white]", p.color, p.id, s)
	p.game.log.Write([]byte(s2))
}

func (p *Player) GetFirstPlayableCard() (int, *Card) {
	for i, c := range p.cards {
		if c.Played {
			continue
		}

		if p.game.validCard(c) {
			return i, c
		}
	}

	return -1, nil
}

func (p *Player) HasPlayableCards() bool {
	valid := false
	for _, c := range p.cards {
		if c.Played {
			continue
		}

		if p.game.validCard(c) {
			valid = true
			break
		}
	}

	return valid
}

func (p *Player) IsPlayableFor(card *Card) bool {

	for _, c := range p.cards {
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

	for _, c := range p.cards {
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

func (p *Player) selectCard(reverse bool) *Card {
	var focusIdx int = 0
	for i, card := range p.cards {

		if !card.HasFocus() {
			continue
		}
		if reverse {
			i = i - 1
			if i < 0 {
				i = len(p.cards) - 1
			}
		} else {
			i = i + 1
			i = i % len(p.cards)
		}

		focusIdx = i
	}

	//if focusIdx < len(p.cards) {
	p.game.App.SetFocus(p.cards[focusIdx])
	p.selectedCard = focusIdx
	//}

	return p.cards[focusIdx]
}
