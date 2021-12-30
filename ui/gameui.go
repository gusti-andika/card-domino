package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type GameUI struct {
	*tview.Flex
	headView   *tview.Flex
	tailView   *tview.Flex
	statusView *tview.TextView
	playerView *tview.Flex
	players    []*PlayerUI
	log        *LogWindow
	head       *Card
	tail       *Card
	last       *Card
	cardNum    int // number of played card
}

func NewGameUI() *GameUI {

	ui := &GameUI{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		headView:   tview.NewFlex(),
		tailView:   tview.NewFlex(),
		playerView: tview.NewFlex().SetDirection(tview.FlexRow),
		log:        NewLogWindow(),
		statusView: tview.NewTextView().SetDynamicColors(true),
		cardNum:    0,
		players:    make([]*PlayerUI, 0, 5),
	}

	ui.log.SetDynamicColors(true)

	// setup UI & layout
	header := tview.NewFlex()
	header.AddItem(ui.headView, 0, 1, false)
	ui.headView.SetBorder(true).SetTitle("Head[First 3 Cards]")
	header.AddItem(ui.tailView, 0, 1, false)
	ui.tailView.SetBorder(true).SetTitle("Tail[Last 3 Cards]")
	ui.AddItem(header, 0, 1, false)
	ui.AddItem(ui.playerView, 0, 3, false)

	ui.statusView.SetBackgroundColor(tcell.ColorYellow)

	logPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	logPanel.SetBorder(true).SetTitle("Log")
	logPanel.AddItem(ui.log, 0, 1, false)
	logPanel.AddItem(ui.statusView, 1, 1, false)
	header.AddItem(logPanel, 0, 1, false)

	return ui
}

func (ui *GameUI) ClearStatus() {
	ui.statusView.Clear()
}
func (ui *GameUI) GetLog() *LogWindow {
	return ui.log
}

func (ui *GameUI) UpdateStatus(status string) {
	ui.statusView.SetText(status)
}

func (ui *GameUI) Log(s string) {
	ui.log.Write([]byte(s))
}

func (ui *GameUI) AddFirst(card *Card) {
	if ui.head == nil {
		ui.head, ui.tail = card, card
	} else {
		card.prev = ui.tail
		ui.tail.next = card
		ui.tail = card
	}

	ui.updateUI(card)

}

func (ui *GameUI) AddCard(playedCard *Card) (head int, tail int, appended bool) {
	isAppend := false
	switch {
	case ui.GetCardNum() == 0:
		head = playedCard.X
		tail = playedCard.Y

	case ui.GetCardNum() >= 1:
		if ui.head.X == playedCard.X {
			head = playedCard.Y
			playedCard.Flip()
		} else if ui.head.X == playedCard.Y {
			head = playedCard.X
		} else if ui.tail.Y == playedCard.X {
			tail = playedCard.Y
			isAppend = true
		} else if ui.tail.Y == playedCard.Y {
			tail = playedCard.X
			isAppend = true
			playedCard.Flip()
		}
	}

	if isAppend {
		ui.AddFirst(playedCard)
	} else {
		ui.AddLast(playedCard)
	}

	appended = isAppend
	return
}

func (ui *GameUI) AddPlayerUI(pui *PlayerUI) {
	ui.players = append(ui.players, pui)
	ui.playerView.Clear()

	for _, p := range ui.players {
		ui.playerView.AddItem(p, 12, 1, false)
	}
}

func (ui *GameUI) GetPlayerUI(id string) *PlayerUI {
	for i := 0; i < ui.GetItemCount(); i++ {
		if pui, ok := ui.GetItem(i).(*PlayerUI); ok && pui.ID == id {
			return pui
		}
	}

	return nil
}

func (ui *GameUI) AddLast(card *Card) {
	if ui.head == nil {
		ui.head, ui.tail = card, card
	} else {
		old := ui.head
		old.prev = card
		ui.head = card
		ui.head.next = old
	}

	ui.updateUI(card)
}

func (ui *GameUI) updateUI(card *Card) {
	ui.cardNum++
	ui.headView.Clear()
	next := ui.head
	for i := 0; next != nil && i < 3; i++ {
		ui.headView.AddItem(next, 10, 1, false)
		next = next.next
	}

	ui.tailView.Clear()
	prev := ui.tail
	for i := 0; prev.prev != nil && i < 2; i++ {
		prev = prev.prev
	}

	for prev != nil {
		ui.tailView.AddItem(prev, 10, 1, false)
		prev = prev.next
	}

	if ui.last != nil {
		ui.last.ClearHighlight()
	}

	ui.last = card
	ui.last.Highlight()
}

func (ui *GameUI) GetCardNum() int {
	return ui.cardNum
}

func (ui *GameUI) GetHeadVal() int {
	if ui.head != nil {
		return ui.head.X
	}

	return 0
}

func (ui *GameUI) GetTailVal() int {
	if ui.tail != nil {
		return ui.tail.Y
	}

	return 0
}

func (ui *GameUI) ValidCard(card *Card) bool {

	switch {
	case card.Played:
		return false
	case ui.GetCardNum() == 0:
		return true
	case ui.GetCardNum() >= 1:
		if ui.GetHeadVal() == card.X {
			return true
		} else if ui.GetHeadVal() == card.Y {
			return true
		} else if ui.GetTailVal() == card.X {
			return true
		} else if ui.GetTailVal() == card.Y {
			return true
		} else {
			return false
		}
	}

	return false
}
