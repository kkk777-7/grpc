package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"reversi/build"
	"reversi/game"
	"reversi/pkg/protobuf"
)

type Reversi struct {
	sync.RWMutex
	started  bool
	finished bool
	me       *game.Player
	room     *game.Room
	game     *game.Game
}

func NewReversi() *Reversi {
	return &Reversi{}
}

func (r *Reversi) Run() int {
	if err := r.run(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func (r *Reversi) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "Failed to connect to grpc server")
	}
	defer conn.Close()

	err = r.matching(ctx, protobuf.NewMatchingServiceClient(conn))
	if err != nil {
		return err
	}

	r.game = game.NewGame(r.me.Color)

	return r.play(ctx, protobuf.NewGameServiceClient(conn))
}

func (r *Reversi) matching(ctx context.Context, cli protobuf.MatchingServiceClient) error {

	stream, err := cli.JoinRoom(ctx, &protobuf.JoinRoomRequest{})
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	fmt.Println("Requested matching...")

	for {
		resp, err := stream.Recv()
		if err != nil {
			return err
		}

		if resp.GetStatus() == protobuf.JoinRoomResponse_MATCHED {
			r.room = build.Room(resp.GetRoom())
			r.me = build.Player(resp.GetMe())
			fmt.Printf("Matched room_id=%d\n", resp.GetRoom().GetId())
			return nil
		} else if resp.GetStatus() == protobuf.JoinRoomResponse_WAITING {
			fmt.Println("Waiting mathing...")
		}
	}
}

func (r *Reversi) play(ctx context.Context, cli protobuf.GameServiceClient) error {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := cli.Play(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	go func() {
		err := r.send(c, stream)
		if err != nil {
			cancel()
		}
	}()

	err = r.recieve(c, stream)
	if err != nil {
		cancel()
		return err
	}

	return nil
}

func (r *Reversi) send(ctx context.Context, stream protobuf.GameService_PlayClient) error {
	for {
		r.RLock()

		if r.finished {
			// recieve側で終了されたので、send側も終了する
			r.RUnlock()
			return nil
		} else if !r.started {
			// 未開始なので、開始リクエストを送る
			err := stream.Send(&protobuf.PlayRequest{
				RoomId: r.room.ID,
				Player: build.PBPlayer(r.me),
				Action: &protobuf.PlayRequest_Start{
					Start: &protobuf.PlayRequest_StartAction{},
				},
			})
			r.RUnlock()
			if err != nil {
				return err
			}

			for {
				// 相手が開始するまで待機する
				r.RLock()
				if r.started {
					// 開始をrecieveした
					r.RUnlock()
					fmt.Println("READY GO!")
					break
				}
				r.RUnlock()
				fmt.Println("Waiting until opponent player ready")
				time.Sleep(1 * time.Second)
			}
		} else {
			// 対戦中

			r.RUnlock()
			// 手の入力を待機する
			fmt.Print("Input Your Move (ex. A-1):")
			stdin := bufio.NewScanner(os.Stdin)
			stdin.Scan()

			// 入力された手を解析する
			text := stdin.Text()
			x, y, err := parseInput(text)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// 手を打つ
			r.Lock()
			_, err = r.game.Move(x, y, r.me.Color)
			r.Unlock()
			if err != nil {
				fmt.Println(err)
				continue
			}

			go func() {
				// サーバーに手を送る
				err = stream.Send(&protobuf.PlayRequest{
					RoomId: r.room.ID,
					Player: build.PBPlayer(r.me),
					Action: &protobuf.PlayRequest_Move{
						Move: &protobuf.PlayRequest_MoveAction{
							Move: &protobuf.Move{
								X: x,
								Y: y,
							},
						},
					},
				})
				if err != nil {
					fmt.Println(err)
				}
			}()

			// 一度手を打ったら5秒間待機する
			ch := make(chan int)
			go func(ch chan int) {
				fmt.Println("")
				for i := 0; i < 5; i++ {
					fmt.Printf("freezing in %d second.\n", (5 - i))
					time.Sleep(1 * time.Second)
				}
				fmt.Println("")
				ch <- 0
			}(ch)
			<-ch
		}

		select {
		case <-ctx.Done():
			// キャンセルされたので終了する
			return nil
		default:
		}
	}
}

// `A-2`の形式で入力された手を (x, y)=(1,2) の形式に変換する
func parseInput(txt string) (int32, int32, error) {
	ss := strings.Split(txt, "-")
	if len(ss) != 2 {
		return 0, 0, fmt.Errorf("入力が不正です。例：A-1")
	}

	xs := ss[0]
	xrs := []rune(strings.ToUpper(xs))
	x := int32(xrs[0]-rune('A')) + 1

	if x < 1 || 8 < x {
		return 0, 0, fmt.Errorf("入力が不正です。例：A-1")
	}

	ys := ss[1]
	y, err := strconv.ParseInt(ys, 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("入力が不正です。例：A-1")
	}
	if y < 1 || 8 < y {
		return 0, 0, fmt.Errorf("入力が不正です。例：A-1")
	}

	return x, int32(y), nil
}

func (r *Reversi) recieve(ctx context.Context, stream protobuf.GameService_PlayClient) error {
	for {
		// サーバーからのストリーミングを受け取る
		res, err := stream.Recv()
		if err != nil {
			return err
		}

		r.Lock()
		switch res.GetEvent().(type) {
		case *protobuf.PlayResponse_Waiting:
			// 開始待機中
		case *protobuf.PlayResponse_Ready:
			// 開始
			r.started = true
			r.game.Display(r.me.Color)
		case *protobuf.PlayResponse_Move:
			// 手を打たれた
			color := build.Color(res.GetMove().GetPlayer().GetColor())
			if color != r.me.Color {
				move := res.GetMove().GetMove()
				// クライアント側のゲーム情報に反映させる
				r.game.Move(move.GetX(), move.GetY(), color)
				fmt.Print("Input Your Move (ex. A-1):")
			}
		case *protobuf.PlayResponse_Finished:
			// ゲームが終了した
			r.finished = true

			// 勝敗を表示する
			winner := build.Color(res.GetFinished().Winner)
			fmt.Println("")
			if winner == game.None {
				fmt.Println("Draw!")
			} else if winner == r.me.Color {
				fmt.Println("You Win!")
			} else {
				fmt.Println("You Lose!")
			}

			// ループを終了する
			r.Unlock()
			return nil
		}
		r.Unlock()

		select {
		case <-ctx.Done():
			// キャンセルされたので終了する
			return nil
		default:
		}
	}
}
