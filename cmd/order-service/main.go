package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/ArsikIT/generated-proto-go/proto/order/v1"
	paymentv1 "github.com/ArsikIT/generated-proto-go/proto/payment/v1"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"order-service/internal/app"
	"order-service/internal/repository/postgres"
	grpctransport "order-service/internal/transport/grpc"
	httptransport "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	cfg := app.LoadConfig()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	grpcConn, err := grpc.NewClient(
		cfg.PaymentGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connect payment grpc: %v", err)
	}
	defer grpcConn.Close()

	orderRepo := postgres.NewOrderRepository(db)
	orderUpdatesBroker := app.NewOrderUpdatesBroker()
	paymentClient := app.NewPaymentGRPCClient(paymentv1.NewPaymentServiceClient(grpcConn))

	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentClient, orderUpdatesBroker)
	handler := httptransport.NewHandler(orderUseCase)
	router := httptransport.NewRouter(handler)
	orderGRPCHandler := grpctransport.NewHandler(orderUseCase, orderUpdatesBroker)

	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, orderGRPCHandler)

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}
	defer grpcListener.Close()

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("order-service listening on %s", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	go func() {
		log.Printf("order-service grpc listening on %s", cfg.GRPCAddress)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
}
