package build

import (
	"fmt"
	"reversi/game"
	"reversi/pkg/protobuf"
)

func Room(r *protobuf.Room) *game.Room {
	return &game.Room{
		ID:    r.GetId(),
		Host:  Player(r.GetHost()),
		Guest: Player(r.GetGuest()),
	}
}

func Player(p *protobuf.Player) *game.Player {
	return &game.Player{
		ID:    p.GetId(),
		Color: Color(p.GetColor()),
	}
}

func Color(c protobuf.Color) game.Color {
	switch c {
	case protobuf.Color_BLACK:
		return game.Black
	case protobuf.Color_WHITE:
		return game.White
	case protobuf.Color_WALL:
		return game.Wall
	case protobuf.Color_EMPTY:
		return game.Empty
	}
	panic(fmt.Sprintf("unknown color=%v", c))
}
