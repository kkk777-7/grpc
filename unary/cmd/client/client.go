package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"pancake/maker/pkg/test"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	scanner *bufio.Scanner
	client  test.PancakeBakerServiceClient
)

func main() {
	fmt.Println("start gRPC Client.")

	scanner = bufio.NewScanner(os.Stdin)

	address := "localhost:50051"
	conn, err := grpc.Dial(
		address,

		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatal("Connection failed.")
		return
	}
	defer conn.Close()

	client = test.NewPancakeBakerServiceClient(conn)

	for {
		fmt.Println("*****************************")
		fmt.Println("1: send Bake Request")
		fmt.Println("2: send Report Request")
		fmt.Println("3: exit")
		fmt.Println("*****************************")
		fmt.Print("please enter >")

		scanner.Scan()
		in := scanner.Text()

		switch in {
		case "1":
			Bake()
		case "2":
			Report()
		case "3":
			fmt.Println("bye.")
			goto M
		}
	}
M:
}

func Bake() {
	fmt.Println("*****************************")
	fmt.Println("Please enter Pancake menu.")
	fmt.Println("1: Pancake_CLASSIC")
	fmt.Println("2: Pancake_BANANA_AND_WHIP")
	fmt.Println("3: Pancake_BACON_AND_CHEESE")
	fmt.Println("4: Pancake_MIX_BERRY")
	fmt.Println("5: Pancake_BAKED_MARSHMALLOW")
	fmt.Println("6: Pancake_SPICY_CURRY")
	fmt.Println("*****************************")
	fmt.Print("please enter >")

	scanner.Scan()
	menu := scanner.Text()
	menuNum, _ := strconv.Atoi(menu)

	req := &test.BakeRequest{
		Menu: test.Pancake_Menu(menuNum),
	}
	res, err := client.Bake(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetPancake())
	}
}

func Report() {
	req := &test.ReportRequest{}
	res, err := client.Report(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetReport())
	}
}
