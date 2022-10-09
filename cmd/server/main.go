package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/inconshreveable/log15"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	"github.com/scalalang2/cosmfaucet/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
)

var (
	logger = log15.New("module", "app")
)

func main() {
	if err := run("localhost:9090"); err != nil {
		logger.Error("failed to run the server", "err", err)
		panic(err)
	}
}

func run(endpoint string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "0.0.0.0:9090"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// insecure tls option
	serverOpts := []grpc.ServerOption{grpc.Creds(insecure.NewCredentials())}
	s := grpc.NewServer(serverOpts...)
	faucetpb.RegisterFaucetServiceServer(s, server.New())

	// run gRPC server
	go func() {
		logger.Info("start to serve gRPC Server", "addr", addr)
		if grpcErr := s.Serve(lis); grpcErr != nil {
			logger.Error("failed to serve grpc server", "err", grpcErr)
			panic(grpcErr)
		}
	}()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = faucetpb.RegisterFaucetServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	logger.Info("start to serve HTTP Proxy Server", "addr", ":8081")
	return http.ListenAndServe(":8081", mux)
}
