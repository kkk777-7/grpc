package game

import "fmt"

type Board struct {
	Cells [][]Color
}

func NewBoard() *Board {
	b := &Board{
		Cells: make([][]Color, 10),
	}
	for i := 0; i < 10; i++ {
		b.Cells[i] = make([]Color, 10)
	}

	for i := 0; i < 10; i++ {
		b.Cells[0][i] = Wall
	}
	for i := 1; i < 9; i++ {
		b.Cells[i][0] = Wall
		b.Cells[i][9] = Wall
	}
	for i := 0; i < 9; i++ {
		b.Cells[9][i] = Wall
	}

	b.Cells[4][4] = White
	b.Cells[5][5] = White
	b.Cells[5][4] = Black
	b.Cells[4][5] = Black

	return b
}

func (b *Board) PutStone(x, y int32, c Color) error {
	if !b.CanPutStone(x, y, c) {
		return fmt.Errorf("can not put stone x=%v, y=%v, color=%v", x, y, ColorToStr(c))
	}

	b.Cells[x][y] = c

	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if b.CountTurnableStonesByDirection(x, y, c, int32(dx), int32(dy)) > 0 {
				b.TurnStonesByDirection(x, y, c, int32(dx), int32(dy))
			}
		}
	}
	return nil
}

func (b *Board) CanPutStone(x, y int32, c Color) bool {
	if b.Cells[x][y] != Empty {
		return false
	}
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if b.CountTurnableStonesByDirection(x, y, c, int32(dx), int32(dy)) > 0 {
				return true
			}
		}
	}
	return false
}

func (b *Board) CountTurnableStonesByDirection(x, y int32, c Color, dx, dy int32) int {
	cnt := 0

	nx := x + dx
	ny := y + dy
	for {
		nc := b.Cells[nx][ny]
		if nc != OpponentColor(c) {
			break
		}
		cnt++

		nx += dx
		ny += dy
	}
	if cnt > 0 && b.Cells[nx][ny] == c {
		return cnt
	}
	return 0
}

func (b *Board) TurnStonesByDirection(x, y int32, c Color, dx, dy int32) {
	nx := x + dx
	ny := y + dy

	for {
		nc := b.Cells[nx][ny]
		if nc != OpponentColor(c) {
			break
		}
		b.Cells[nx][ny] = c

		nx += dx
		ny += dy
	}
}

func (b *Board) AvailableCellCount(c Color) int {
	cnt := 0

	for i := 1; i < 9; i++ {
		for j := 1; j < 9; j++ {
			if b.CanPutStone(int32(i), int32(j), c) {
				cnt++
			}
		}
	}
	return cnt
}

func (b *Board) Score(c Color) int {
	cnt := 0

	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if b.Cells[i][j] != c {
				continue
			}
			cnt++
		}
	}
	return cnt
}

func (b *Board) Rest() int {
	cnt := 0

	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if b.Cells[i][j] == Empty {
				cnt++
			}
		}
	}
	return cnt
}
