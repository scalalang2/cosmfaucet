package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	app.logger = logger
	defer logger.Sync()

	// connects to all chains
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	keyDir := wd + "/keys"

	for _, chain := range app.config.Chains {
		fields := []zap.Field{zap.String("chain", chain.Name), zap.String("rpc", chain.RpcEndpoint)}
		logger.Info("trying to connect to the chain", fields...)
		cc, err := newChainClient(logger, chain, keyDir)
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

// validateChainConditions check the account can conduct a duty of faucet.
// It must have enough balance to pay out the drop coins to the user
func (a *App) validateChainConditions() error {
	for _, chain := range a.config.Chains {
		valid := false
		client, ok := a.clients[chain.ChainId]
		if !ok {
			return fmt.Errorf("chain with id %s does not exist", chain.ChainId)
		}

		addr, err := sdk.GetFromBech32(chain.Sender, chain.AccountPrefix)
		if err != nil {
			return err
		}

		dropCoin, err := sdk.ParseCoinNormalized(chain.DropCoin)
		if err != nil {
			return err
		}

		coins, err := client.QueryBalanceWithDenomTraces(context.Background(), addr, nil)
		if err != nil {
			return err
		}

		for _, coin := range coins {
			if coin.Denom == dropCoin.Denom {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("chain %s doesnt have the valid denom of drop coin: %s", chain.Name, dropCoin.Denom)
		}
	}

	return nil
}

func (a *App) Run() error {
	err := a.validateChainConditions()
	if err != nil {
		return fmt.Errorf("validation check failed on chains: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel

	grpcAddr := fmt.Sprintf("0.0.0.0:%d", a.config.Server.Grpc.Port)
	httpAddr := fmt.Sprintf(":%d", a.config.Server.Http.Port)

	// run gRPC server
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

	// run HTTP Server
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

// newChainClient create a client for the cosmos blockchain
func newChainClient(logger *zap.Logger, config ChainConfig, homePath string) (*lens.ChainClient, error) {
	if !strings.HasPrefix(config.RpcEndpoint, "http") {
		return nil, errInvalidEndpoint{rpc: config.RpcEndpoint}
	}

	cfg := lens.ChainClientConfig{
		Key:            "default",
		ChainID:        config.ChainId,
		RPCAddr:        config.RpcEndpoint,
		AccountPrefix:  config.AccountPrefix,
		KeyringBackend: "test",
		GasAdjustment:  config.GasAdjustment,
		GasPrices:      config.GasPrice,
		KeyDirectory:   homePath,
		Debug:          false,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lens.ModuleBasics,
	}

	cc, err := lens.NewChainClient(logger, &cfg, homePath, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	return cc, nil
}
