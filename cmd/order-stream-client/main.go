package main

import (
	"context"
	"fmt"
	"log"
	"os"

	orderv1 "github.com/ArsikIT/generated-proto-go/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/order-stream-client <order-id>")
	}

	addr := os.Getenv("ORDER_GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:50052"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect order grpc: %v", err)
	}
	defer conn.Close()

	client := orderv1.NewOrderServiceClient(conn)
	stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderv1.OrderRequest{
		OrderId: os.Args[1],
	})
	if err != nil {
		log.Fatalf("subscribe: %v", err)
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("recv: %v", err)
		}

		fmt.Printf("order=%s status=%s updated_at=%s\n", update.GetOrderId(), update.GetStatus(), update.GetUpdatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00"))
	}
}
