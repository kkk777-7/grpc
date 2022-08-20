package handler

import (
	"fmt"
	"sync"

	"reversi/build"
	"reversi/game"
	"reversi/pkg/protobuf"
)

type GameHandler struct {
	protobuf.UnimplementedGameServiceServer
	sync.RWMutex
	games  map[int32]*game.Game
	client map[int32][]protobuf.GameService_PlayServer
}

func NewGameHandler() *GameHandler {
	return &GameHandler{
		games:  make(map[int32]*game.Game),
		client: make(map[int32][]protobuf.GameService_PlayServer),
	}
}

func (h *GameHandler) Play(stream protobuf.GameService_PlayServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		roomID := req.GetRoomId()
		player := build.Player(req.GetPlayer())

		switch req.GetAction().(type) {
		case *protobuf.PlayRequest_Start:
			err := h.start(stream, roomID, player)
			if err != nil {
				return err
			}
		case *protobuf.PlayRequest_Move:
			action := req.GetMove()
			x := action.GetMove().GetX()
			y := action.GetMove().GetY()
			err := h.move(roomID, x, y, player)
			if err != nil {
				return err
			}
		}
	}
}

func (h *GameHandler) start(stream protobuf.GameService_PlayServer, roomID int32, me *game.Player) error {
	h.Lock()
	defer h.Unlock()

	g := h.games[roomID]
	if g == nil {
		g = game.NewGame(game.None)
		h.games[roomID] = g
		h.client[roomID] = make([]protobuf.GameService_PlayServer, 0, 2)
	}

	h.client[roomID] = append(h.client[roomID], stream)

	if len(h.client[roomID]) == 2 {
		for _, s := range h.client[roomID] {
			err := s.Send(&protobuf.PlayResponse{
				Event: &protobuf.PlayResponse_Ready{
					Ready: &protobuf.PlayResponse_ReadyEvent{},
				},
			})
			if err != nil {
				return err
			}
		}
		fmt.Printf("game has started room_id=%v\n", roomID)
	} else {
		err := stream.Send(&protobuf.PlayResponse{
			Event: &protobuf.PlayResponse_Waiting{
				Waiting: &protobuf.PlayResponse_WaitingEvent{},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *GameHandler) move(roomID int32, x int32, y int32, p *game.Player) error {
	h.Lock()
	defer h.Unlock()

	g := h.games[roomID]

	finished, err := g.Move(x, y, p.Color)
	if err != nil {
		return err
	}

	for _, s := range h.client[roomID] {
		err := s.Send(&protobuf.PlayResponse{
			Event: &protobuf.PlayResponse_Move{
				Move: &protobuf.PlayResponse_MoveEvent{
					Player: build.PBPlayer(p),
					Move: &protobuf.Move{
						X: x,
						Y: y,
					},
					Board: build.PBBoard(g.Board),
				},
			},
		})
		if err != nil {
			return err
		}

		if finished {
			err := s.Send(
				&protobuf.PlayResponse{
					Event: &protobuf.PlayResponse_Finished{
						Finished: &protobuf.PlayResponse_FinishedEvent{
							Winner: build.PBColor(g.Winner()),
							Board:  build.PBBoard(g.Board),
						},
					},
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
