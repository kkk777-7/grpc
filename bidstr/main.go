package main

import (
	"fmt"
	"reversi/game"
)

func main() {
	new := game.NewGame(game.Black)
	ok, err := new.Move(4, 3, game.White)
	fmt.Println(ok, err)
}
