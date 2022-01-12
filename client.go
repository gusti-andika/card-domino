package domino

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/gusti-andika/domino/rpc"
	"github.com/gusti-andika/domino/ui"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
)

type Client struct {
	*ui.GameUI
	app *tview.Application
	//stream        rpc.GameService_UpdateClient
	rpcClient     rpc.GameServiceClient
	conn          *grpc.ClientConn
	mePlayerId    string
	players       []*Player
	currentPlayer bool
}

func NewClient(playerName string) *Client {
	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start client :%v", err)
	}

	clientRpc := rpc.NewGameServiceClient(clientConn)
	res, err := clientRpc.Join(context.Background(), &rpc.JoinRequest{PlayerName: playerName})

	if err != nil {
		log.Fatalf("Failed to start client :%v", err)
	}

	player := NewClientPlayer(playerName)
	player.SetId(res.GetPlayer().GetId())
	player.SetColor(res.GetPlayer().GetColor())

	client := &Client{
		GameUI:     ui.NewGameUI(),
		conn:       clientConn,
		app:        tview.NewApplication(),
		rpcClient:  clientRpc,
		mePlayerId: player.Id,
		players:    make([]*Player, 0, 5),
	}

	client.AddPlayer(player)
	for _, opponent := range res.GetPlayersInGame() {
		client.handleResultOpponent(opponent)
	}

	client.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !client.currentPlayer {
			return event
		}

		switch event.Key() {
		case tcell.KeyRight:
			player.navigateRight()
			client.app.SetFocus(player.GetSelectedCard())
		case tcell.KeyLeft:
			player.navigateLeft()
			client.app.SetFocus(player.GetSelectedCard())
		case tcell.KeyEnter:
			client.handlePlayerMove()
		}

		return event
	})

	return client
}

func (client *Client) Run() {
	go handleMessage(client)
	if err := client.app.SetRoot(client, true).Run(); err != nil {
		panic(err)
	}
}

func (client *Client) AddPlayer(p *Player) {
	client.players = append(client.players, p)
	client.AddPlayerUI(p.ui)
}

func logg(c *Client, msg string) {
	c.app.QueueUpdateDraw(func() {
		c.Log(msg + "\n")
	})
}

func handleMessage(client *Client) {
	stream, err := client.rpcClient.Update(context.Background())
	if err != nil {
		log.Fatalf("Failed to start client :%v", err)
	}

	me := client.getMePlayer()
	err = stream.SendMsg(&rpc.GameUpdate{
		Cmd: "getcards",
		Instrument: &rpc.GameUpdate_Player{
			Player: &rpc.Player{
				Id: me.Id,
			},
		},
	})

	if err != nil {
		logg(client, fmt.Sprintf("%#v", err))
		return
	}

	go func() {
		me := client.getMePlayer()
		for out := range me.out {
			logg(client, fmt.Sprintf("send: %+v", out))
			if err := stream.SendMsg(out); err != nil {
				logg(client, fmt.Sprintf("%+v", err))
			}
		}
	}()

	for {

		res, err := stream.Recv()
		logg(client, fmt.Sprintf("recv: %+v", res))
		if err == io.EOF {
			logg(client, fmt.Sprintf("EOF: %#v", err))
			break
		}

		if err != nil {
			logg(client, fmt.Sprintf("%#v", err))
			break
		}

		switch res.Instrument.(type) {

		case *rpc.GameUpdate_Player:
			switch res.Cmd {
			case "newplayer":
				client.app.QueueUpdateDraw(func() {
					client.handleResultOpponent(res.GetPlayer())
				})
			case "playerturn":
				client.handleResultPlayerTurn(res.GetPlayer())
			default:
				client.handleResultPlayer(res.GetPlayer())
			}

		case *rpc.GameUpdate_Move:
			client.handleResultMove(res.GetMove())
		case *rpc.GameUpdate_Log:
			client.handleResultLog(res.GetLog())
		case *rpc.GameUpdate_GetCards:
			client.handleResultGetCards(res.GetGetCards())
		case *rpc.GameUpdate_InitialCard:
			client.handleResultInitialCard(res.GetInitialCard())
		}

	}
}

func (c *Client) handleResultPlayer(p *rpc.Player) {
	player := NewPlayer(p.GetName())
	player.Id = p.GetId()

	cards := make([]*ui.Card, len(p.GetCards()))
	for i, v := range p.GetCards() {
		cards[i] = &ui.Card{
			X: int(v.X),
			Y: int(v.Y),
		}
	}

	player.ui.SetCards(cards)
	c.AddPlayerUI(player.ui)
}

func (c *Client) getMePlayer() *Player {
	for _, p := range c.players {
		if c.mePlayerId == p.Id {
			return p
		}
	}

	return nil
}

func (c *Client) getPlayer(id string) *Player {
	for _, p := range c.players {
		if id == p.Id {
			return p
		}
	}

	return nil
}

func (c *Client) handleResultPlayerTurn(p *rpc.Player) {
	c.currentPlayer = c.mePlayerId == p.Id
	for _, other := range c.players {
		if other.Id == p.Id {
			c.app.SetFocus(other.ui)
			return
		}
	}
}

func (c *Client) handleResultOpponent(p *rpc.Player) *Player {
	player := NewClientOpponent(p.GetName(), p.GetId(), p.GetColor())

	cards := make([]*ui.Card, 5)
	for i, _ := range cards {
		cards[i] = ui.NewCard(-1, -1)
		cards[i].SetHideNotPlayedCards(true)
	}

	player.ui.SetCards(cards)
	c.AddPlayer(player)
	return player
}

func (c *Client) handleResultMove(r *rpc.Move) {

	for _, p := range c.players {
		if p.Id == r.Player.Id {
			card := p.ui.GetCard(int(r.CardIndex))
			card.SetValue(int(r.Card.X), int(r.Card.Y))
			p.ui.PlayCard(int(r.CardIndex))
			c.app.QueueUpdateDraw(func() {
				c.AddCard(ui.NewCard(card.X, card.Y))
			})

		}

		if p.Id == r.NextPlayer.Id {
			c.handleResultPlayerTurn(r.NextPlayer)
		}
	}

}

func (c *Client) handlePlayerMove() {
	mePlayer := c.getMePlayer()
	card := mePlayer.GetSelectedCard()

	msg := &rpc.GameUpdate{
		Cmd: "playermove",
		Instrument: &rpc.GameUpdate_Move{
			Move: &rpc.Move{
				Card:      &rpc.Card{X: int32(card.X), Y: int32(card.Y)},
				CardIndex: int32(mePlayer.GetSelectedCardIndex()),
				Player:    &rpc.Player{Id: mePlayer.Id},
			},
		},
	}

	mePlayer.out <- msg
}

func (c *Client) handleResultLog(r string) {
	c.Log(r)
}

func (c *Client) handleResultGetCards(cards *rpc.GetCards) {
	clientCards := make([]*ui.Card, len(cards.Cards))
	for i, c := range cards.Cards {
		clientCards[i] = ui.NewCard(int(c.X), int(c.Y))
	}

	mePlayer := c.getMePlayer()
	c.app.QueueUpdateDraw(func() {
		mePlayer.ui.SetCards(clientCards)
	})
}

func (c *Client) handleResultInitialCard(card *rpc.Card) {
	initialCard := ui.NewCard(int(card.X), int(card.Y))
	initialCard.Play()
	c.app.QueueUpdateDraw(func() {
		c.GameUI.AddCard(initialCard)
	})

}
