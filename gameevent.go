package domino

import (
	"fmt"
	"log"

	"github.com/gusti-andika/domino/eventbus"
	"github.com/gusti-andika/domino/rpc"
	"github.com/gusti-andika/domino/ui"
)

type gameEvent struct {
	game *Game
}

func InitGameEventHandlers(game *Game) *gameEvent {
	eventHandlers := &gameEvent{game: game}
	eventHandlers.initHandlers()
	return eventHandlers
}

func (ge *gameEvent) initHandlers() {
	eventbus.AddHandler(eventbus.PlayerJoined, ge.notifyNewPlayer)
	eventbus.AddHandler(eventbus.InitialCard, ge.notifyInitialCard)
	eventbus.AddHandler(eventbus.InvalidMove, ge.notifyInvalidMove)
	eventbus.AddHandler(eventbus.SkipTurn, ge.notifySkipTurn)
	eventbus.AddHandler(eventbus.PlayerTurn, ge.notifyPlayerTurn)
	eventbus.AddHandler(eventbus.PlayerMoved, ge.notifyPlayerMove)
	eventbus.AddHandler(eventbus.CardAssigned, ge.notifyCardAssigned)
	eventbus.AddHandler(eventbus.GameFinished, ge.notifyGameFinished)
	eventbus.AddHandler(eventbus.GameLog, ge.notifyGameLog)
}

func (ge *gameEvent) notifyNewPlayer(event eventbus.Event) {
	switch newPlayer := event.Data["player"].(type) {
	case *Player:
		for _, p := range ge.game.Players {
			if p.Id == newPlayer.Id {
				continue
			}

			msg := &rpc.GameUpdate{
				Cmd: "newplayer",
				Instrument: &rpc.GameUpdate_Player{
					Player: &rpc.Player{
						Id:    newPlayer.Id,
						Name:  newPlayer.Name,
						Color: newPlayer.color,
					},
				},
			}

			p.out <- msg
		}

	default:
		log.Printf("Invalid data for event %v -> %v\n", event.Type, event.Data)
	}

}

func (ge *gameEvent) notifyPlayerMove(event eventbus.Event) {
	current := event.Data["currentplayer"].(*Player)
	next := event.Data["nextplayer"].(*Player)
	card := event.Data["card"].(*ui.Card)
	for _, p := range ge.game.Players {
		msg := &rpc.GameUpdate{
			Cmd: "playermove",
			Instrument: &rpc.GameUpdate_Move{
				Move: &rpc.Move{
					Player:     &rpc.Player{Id: current.Id},
					NextPlayer: &rpc.Player{Id: next.Id},
					Card:       &rpc.Card{X: int32(card.X), Y: int32(card.Y)},
				},
			},
		}

		p.out <- msg

		//ge.game.Log(fmt.Sprintf("send move to player : %+v", p))
	}
}

func (ge *gameEvent) notifyCardAssigned(event eventbus.Event) {
	ge.game.Log(fmt.Sprintf("card assigned: %+v", event.Data["cards"]))
	playerId := event.Data["playerId"].(string)

	if player := ge.game.GetPlayerById(playerId); player != nil {
		rpcCards := make([]*rpc.Card, player.ui.GetCardNum())
		for i := 0; i < player.ui.GetCardNum(); i++ {
			c := player.ui.GetCard(i)
			rpcCards[i] = &rpc.Card{X: int32(c.X), Y: int32(c.Y)}
		}

		msg := &rpc.GameUpdate{
			Cmd: "0",
			Instrument: &rpc.GameUpdate_GetCards{
				GetCards: &rpc.GetCards{
					Player: &rpc.Player{Id: playerId},
					Cards:  rpcCards,
				},
			},
		}
		ge.game.Log(fmt.Sprintf("send cards to player : %s, cards: %+v", player.Name, rpcCards))
		player.out <- msg
	}

	ge.game.Log(fmt.Sprintf("num player : %d", len(ge.game.Players)))
	if len(ge.game.Players) >= 2 {
		ge.game.App.QueueUpdateDraw(func() {
			ge.game.start()
		})
	}
}

func (ge *gameEvent) notifyGameFinished(event eventbus.Event) {

}

func (ge *gameEvent) notifyInitialCard(event eventbus.Event) {
	for _, p := range ge.game.Players {
		initialCard := event.Data["card"].(*ui.Card)
		msg := &rpc.GameUpdate{
			Cmd: "initialcard",
			Instrument: &rpc.GameUpdate_InitialCard{
				InitialCard: &rpc.Card{
					X: int32(initialCard.X), Y: int32(initialCard.Y),
				},
			},
		}

		p.out <- msg
	}
}

func (ge *gameEvent) notifyInvalidMove(event eventbus.Event) {

}

func (ge *gameEvent) notifySkipTurn(event eventbus.Event) {

}

func (ge *gameEvent) notifyPlayerTurn(event eventbus.Event) {
	player := event.Data["player"].(*Player)
	for _, p := range ge.game.Players {
		p.out <- &rpc.GameUpdate{
			Cmd: "playerturn",
			Instrument: &rpc.GameUpdate_Player{
				Player: &rpc.Player{
					Id: player.Id,
				},
			},
		}
	}

}

func (ge *gameEvent) notifyGameLog(event eventbus.Event) {
	var msg string
	var ok bool
	if msg, ok = event.Data["msg"].(string); !ok {
		log.Print("Invalid message for log event")
		//return
	}

	ge.game.App.QueueUpdateDraw(func() {
		ge.game.ui.Log(msg)
	})

	// for _, p := range ge.game.Players {
	// 	if p.ClientChannel != nil {
	// 		p.ClientChannel.Send(&rpc.GameUpdate{
	// 			Cmd:        "log",
	// 			Instrument: &rpc.GameUpdate_Log{Log: msg},
	// 		})
	// 	}
	// }

}
