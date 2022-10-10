package core

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	net "net"
	"net/http"
	"os"
)

type ChainId = string
type ChainClients = map[ChainId]*lens.ChainClient

type App struct {
	logger  *zap.Logger
	Server  *Server
	config  *RootConfig
	clients ChainClients

	cancelFunc context.CancelFunc
}

func NewApp(config *RootConfig) (*App, error) {
	app := &App{config: config, clients: make(ChainClients)}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	keyDir := wd + "/keys"

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	app.logger = logger
	defer logger.Sync()

	// connects to all chains
	for _, chain := range app.config.Chains {
		cfg := lens.ChainClientConfig{
			Key:            "default",
			ChainID:        chain.ChainId,
			RPCAddr:        chain.RpcEndpoint,
			AccountPrefix:  chain.AccountPrefix,
			KeyringBackend: "test",
			GasAdjustment:  chain.GasAdjustment,
			GasPrices:      chain.GasPrice,
			KeyDirectory:   keyDir,
			Debug:          false,
			Timeout:        "20s",
			OutputFormat:   "json",
			SignModeStr:    "direct",
			Modules:        lens.ModuleBasics,
		}

		fields := []zap.Field{zap.String("chain", chain.Name), zap.String("rpc", chain.RpcEndpoint)}
		logger.Info("trying to connect to the chain", fields...)
		cc, err := lens.NewChainClient(logger, &cfg, keyDir, os.Stdin, os.Stdout)
		if err != nil {
			logger.Fatal("failed to connect to the chain", fields...)
			return nil, err
		}

		_, ok := app.clients[chain.ChainId]
		if ok {
			return nil, fmt.Errorf("chain with id %s already exists", chain.ChainId)
		}
		app.clients[chain.ChainId] = cc
	}

	app.Server = NewServer(app.config, &app.clients)

	return app, nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel

	grpcAddr := fmt.Sprintf("0.0.0.0:%d", a.config.Server.Grpc.Port)
	httpAddr := fmt.Sprintf(":%d", a.config.Server.Http.Port)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}

	serverOpts := []grpc.ServerOption{grpc.Creds(insecure.NewCredentials())}
	sv := grpc.NewServer(serverOpts...)
	faucetpb.RegisterFaucetServiceServer(sv, a.Server)

	go func() {
		a.logger.Info("start to serve gRPC Server", zap.String("addr", grpcAddr))
		if err := sv.Serve(lis); err != nil {
			a.logger.Fatal("failed to serve gRPC Server", zap.Error(err))
			a.cancelFunc()
		}
	}()

	mux := runtime.NewServeMux()
	endpoint := fmt.Sprintf("localhost:%d", a.config.Server.Grpc.Port)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = faucetpb.RegisterFaucetServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	}

	httpSv := http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	a.logger.Info("start to serve http Server", zap.String("addr", httpAddr))
	return httpSv.ListenAndServe()
}

func (a *App) Stop() {
	a.cancelFunc()
}
