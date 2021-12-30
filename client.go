package domino

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/gusti-andika/domino/rpc"
	"github.com/gusti-andika/domino/ui"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
)

type Client struct {
	*ui.GameUI
	app          *tview.Application
	stream       rpc.GameService_UpdateClient
	rpcClient    rpc.GameServiceClient
	conn         *grpc.ClientConn
	mePlayer     *Player
	otherPlayers []*Player
}

func NewClient(playerName string) *Client {
	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start client :%v", err)
	}

	clientRpc := rpc.NewGameServiceClient(clientConn)
	res, err := clientRpc.Join(context.Background(), &rpc.JoinRequest{PlayerName: playerName})
	//log.Fatalf("res: %+v", res)

	if err != nil {
		log.Fatalf("Failed to start client :%v", err)
	}

	player := NewPlayer(playerName)
	player.Id = res.Player.Id
	player.color = res.Player.Color

	client := &Client{
		GameUI: ui.NewGameUI(),
		conn:   clientConn,
		app:    tview.NewApplication(),
		//stream: stream,
		rpcClient: clientRpc,
		mePlayer:  player,
	}

	//client.app.QueueUpdate(func() {
	client.Log(fmt.Sprintf("%+v", res))
	client.AddPlayerUI(player.ui)

	for _, opponent := range res.GetPlayersInGame() {
		client.handleResultOpponent(opponent)
	}
	//})

	return client
}

func (client *Client) Run() {
	go handleMessage(client)
	if err := client.app.SetRoot(client, true).Run(); err != nil {
		panic(err)
	}
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

	err = stream.SendMsg(&rpc.GameUpdate{
		Cmd: "getcards",
		Instrument: &rpc.GameUpdate_Player{
			Player: &rpc.Player{
				Id: client.mePlayer.Id,
			},
		},
	})

	if err != nil {
		logg(client, fmt.Sprintf("%#v", err))
		return
	}

	for {

		res, err := stream.Recv()
		logg(client, fmt.Sprintf("hadle message5 : %+v", res))
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
			if res.Cmd == "newplayer" {
				client.app.QueueUpdateDraw(func() {
					client.handleResultOpponent(res.GetPlayer())
				})

			} else {
				client.handleResultPlayer(res.GetPlayer())
			}
		case *rpc.GameUpdate_Move:
			client.handleResultMove(res.GetMove())
		case *rpc.GameUpdate_Log:
			client.handleResultLog(res.GetLog())
		case *rpc.GameUpdate_GetCards:
			client.handleResultGetCards(res.GetGetCards())

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
func (c *Client) handleResultOpponent(p *rpc.Player) {
	player := NewOpponent(p.GetName(), p.GetId(), p.GetColor())

	cards := make([]*ui.Card, 5)
	for i, _ := range cards {
		cards[i] = ui.NewCard(-1, -1)
		cards[i].SetHideNotPlayedCards(true)
	}

	player.ui.SetCards(cards)
	c.Log(fmt.Sprintf("%+v", *player.ui))
	c.AddPlayerUI(player.ui)
}

func (c *Client) handleResultMove(r *rpc.Move) {
	pui := c.GetPlayerUI(r.Player.Id)
	if pui == nil {
		return
	}

	if pui.ID != c.mePlayer.Id {
		pui.AddCard(ui.NewCard(int(r.Card.X), int(r.Card.Y)))
	}
}

func (c *Client) handleResultLog(r string) {
	c.Log(r)
}

func (c *Client) handleResultGetCards(cards *rpc.GetCards) {
	clientCards := make([]*ui.Card, len(cards.Cards))
	for i, c := range cards.Cards {
		clientCards[i] = ui.NewCard(int(c.X), int(c.Y))
	}

	c.app.QueueUpdateDraw(func() {
		c.mePlayer.ui.SetCards(clientCards)
	})
}
