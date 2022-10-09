package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	"github.com/scalalang2/cosmfaucet/server/faucet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
)

func main() {
	if err := run("localhost:9090"); err != nil {
		panic(err)
	}
}

func run(endpoint string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := "0.0.0.0:9090"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// insecure tls option
	serverOpts := []grpc.ServerOption{grpc.Creds(insecure.NewCredentials())}
	s := grpc.NewServer(serverOpts...)
	faucetpb.RegisterFaucetServiceServer(s, faucet.New())

	// run gRPC server
	go func() {
		if grpcErr := s.Serve(lis); grpcErr != nil {
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
	return http.ListenAndServe(":8081", mux)
}
