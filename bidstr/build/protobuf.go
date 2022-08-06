package build

import (
	"reversi/game"
	"reversi/pkg/protobuf"
)

func PBRoom(r *game.Room) *protobuf.Room {
	return &protobuf.Room{
		Id:    r.ID,
		Host:  PBPlayer(r.Host),
		Guest: PBPlayer(r.Guest),
	}
}

func PBPlayer(p *game.Player) *protobuf.Player {
	if p == nil {
		return nil
	}
	return &protobuf.Player{
		Id:    p.ID,
		Color: PBColor(p.Color),
	}
}

func PBColor(c game.Color) protobuf.Color {
	switch c {
	case game.Black:
		return protobuf.Color_BLACK
	case game.White:
		return protobuf.Color_WHITE
	case game.Empty:
		return protobuf.Color_EMPTY
	case game.Wall:
		return protobuf.Color_WALL
	}
	return protobuf.Color_UNKNOWN
}

func PBBoard(b *game.Board) *protobuf.Board {
	pbCols := make([]*protobuf.Board_Col, 0, 10)

	for _, col := range b.Cells {
		pbCells := make([]protobuf.Color, 0, 10)
		for _, c := range col {
			pbCells = append(pbCells, PBColor(c))
		}
		pbCols = append(pbCols, &protobuf.Board_Col{
			Cells: pbCells,
		})
	}

	return &protobuf.Board{
		Cols: pbCols,
	}
}
