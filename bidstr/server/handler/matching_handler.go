package handler

import (
	"context"
	"fmt"
	"reversi/build"
	"reversi/game"
	"reversi/pkg/protobuf"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchingHandler struct {
	protobuf.UnimplementedMatchingServiceServer
	sync.RWMutex
	Rooms       map[int32]*game.Room
	maxPlayerID int32
}

func NewMatchingHandler() *MatchingHandler {
	return &MatchingHandler{
		Rooms: make(map[int32]*game.Room),
	}
}

func (h *MatchingHandler) JoinRoom(req *protobuf.JoinRoomRequest, stream protobuf.MatchingService_JoinRoomServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), 2*time.Minute)
	defer cancel()

	h.Lock()

	h.maxPlayerID++
	me := &game.Player{
		ID: h.maxPlayerID,
	}

	for _, room := range h.Rooms {
		if room.Guest == nil {
			me.Color = game.White
			room.Guest = me
			stream.Send(&protobuf.JoinRoomResponse{
				Room:   build.PBRoom(room),
				Me:     build.PBPlayer(me),
				Status: protobuf.JoinRoomResponse_MATCHED,
			})
			h.Unlock()
			fmt.Printf("matched room_id=%v\n", room.ID)
			return nil
		}
	}

	me.Color = game.Black
	room := &game.Room{
		ID:   int32(len(h.Rooms) + 1),
		Host: me,
	}
	h.Rooms[room.ID] = room
	h.Unlock()

	stream.Send(&protobuf.JoinRoomResponse{
		Room:   build.PBRoom(room),
		Status: protobuf.JoinRoomResponse_WAITING,
	})

	ch := make(chan int)
	go func(ch chan<- int) {
		for {
			h.RLock()
			guest := room.Guest
			h.RUnlock()

			if guest != nil {
				stream.Send(&protobuf.JoinRoomResponse{
					Status: protobuf.JoinRoomResponse_MATCHED,
					Room:   build.PBRoom(room),
					Me:     build.PBPlayer(room.Host),
				})
				ch <- 0
				break
			}
			time.Sleep(1 * time.Second)

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}(ch)

	select {
	case <-ch:
	case <-ctx.Done():
		return status.Errorf(codes.DeadlineExceeded, "マッチングできませんでした")
	}

	return nil
}
