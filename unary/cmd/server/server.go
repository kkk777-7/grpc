package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"pancake/maker/cmd/handler"
	"pancake/maker/pkg/test"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	grpc_zap.ReplaceGrpcLogger(zapLogger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_zap.UnaryServerInterceptor(zapLogger),
		),
	)
	test.RegisterPancakeBakerServiceServer(server, handler.NewBakerHandler())
	reflection.Register(server)

	go func() {
		log.Printf("start gRPC server port: %v", port)
		server.Serve(lis)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	server.GracefulStop()
}
