package domino

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/gusti-andika/domino/eventbus"
	"github.com/gusti-andika/domino/rpc"
	"google.golang.org/grpc"
)

type Server struct {
	rpc.UnimplementedGameServiceServer
	game *Game
}

var (
	started = false
	service *Server
)

func (s *Server) Join(_ context.Context, req *rpc.JoinRequest) (res *rpc.JoinResponse, err error) {
	s.game.Log(fmt.Sprintf("received join req: %+v", req))
	player := NewPlayer(req.PlayerName)
	c := make(chan *rpc.JoinResponse)
	go func() {
		s.game.App.QueueUpdateDraw(func() {
			if s.game.Join(player) {
				res := &rpc.JoinResponse{
					Player: &rpc.Player{
						Id:    player.Id,
						Name:  player.Name,
						Color: player.color,
					},
				}
				joinedPlayers := []*rpc.Player{}
				for _, p := range s.game.Players {
					if p.Id == player.Id {
						continue
					}

					joinedPlayers = append(joinedPlayers, &rpc.Player{
						Id: p.Id, Name: p.Name, Color: p.color,
					})
				}

				res.PlayersInGame = joinedPlayers
				c <- res
			}

			close(c)
		})
	}()

	res = <-c
	if res == nil {
		err = fmt.Errorf("failed to join player %s", req.PlayerName)
	}
	return
}

func (s *Server) Update(stream rpc.GameService_UpdateServer) error {
	s.game.Log(fmt.Sprintf("received update: %+v", stream))
	var player *Player
	for {
		// req := rpc.GameChannel{}
		req, err := stream.Recv()
		s.game.Log(fmt.Sprintf("received %+v", req))
		if err == io.EOF {
			return nil
		}

		if err != nil {
			s.game.Log(fmt.Sprintf("%v", err))
			return err
		}

		switch req.Cmd {
		case "getcards":

			player, _ = s.handlePlayerGetCards(stream, req)
			s.game.Log(fmt.Sprintf("call handlePlayerGetCards: %+v", player))

		case "move":
			s.handlePlayerMove(stream, req)
		}

		if player != nil {
			s.game.Log("call go func")

			go func() {
				for {
					select {
					case msg := <-player.out:
						s.game.Log(fmt.Sprintf("msg to send: %+v", msg))
						stream.SendMsg(msg)
					default:
						s.game.Log(fmt.Sprintf("%v no msg to deliver", player.Id))
						time.Sleep(5 * time.Second)
					}
				}
			}()

		}

	}
}

func (s *Server) handlePlayerGetCards(stream rpc.GameService_UpdateServer, req *rpc.GameUpdate) (*Player, bool) {
	switch t := req.Instrument.(type) {

	case *rpc.GameUpdate_Player:
		player := s.game.GetPlayerById(t.Player.Id)
		if player == nil {
			stream.Send(&rpc.GameUpdate{
				Cmd: "",
				Instrument: &rpc.GameUpdate_Log{
					Log: fmt.Sprintf("player %s not exists", t.Player.Id),
				},
			})
			return nil, false
		}

		cards := s.game.Deck.PopCards(5)
		s.game.App.QueueUpdateDraw(func() {
			player.AssignCards(cards)
		})

		return player, true
	default:
		log.Printf("received wrong data for getcards: %v\n", t)
	}
	return nil, false
}

func (s *Server) handlePlayerMove(stream rpc.GameService_UpdateServer, req *rpc.GameUpdate) bool {

	switch t := req.Instrument.(type) {

	case *rpc.GameUpdate_Move:
		currentPlayer := s.game.CurrentPlayer()

		if currentPlayer.Id != t.Move.Player.Id {
			s.game.Log(fmt.Sprintf("Received moved from %s but current player is %s", t.Move.Player.Id, s.game.CurrentPlayer().Id))
			return false
		}

		if !currentPlayer.SelectCard(int(t.Move.CardIndex)) {
			currentPlayer.Log(fmt.Sprintf("Invalid move. Out of range %d ", t.Move.CardIndex))
			return false
		}
		s.game.update()
		return true

	default:
		log.Printf("received wrong data for move: %v\n", t)

	}

	return false
}

func StartServer(game *Game) *Server {
	if started {
		log.Printf("Server already started")
		game.Log("server alrraedy started")

		return service
	}

	listen, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		game.Log(fmt.Sprintf("error start server: %+v", err))
		log.Fatalf("Error start server: %v", err)
	}

	service = &Server{
		game: game,
	}
	grpcServer := grpc.NewServer()
	rpc.RegisterGameServiceServer(grpcServer, service)
	started = true
	eventbus.Start()
	game.Log("server started")

	if err := grpcServer.Serve(listen); err != nil {
		game.Log(fmt.Sprintf("error start server: %+v", err))

		log.Fatalf("Error start server: %v", err)
	}

	return service
}

func ServerIsStarted() bool {
	return started
}
