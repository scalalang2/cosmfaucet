package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChainId = string
type ChainClients = map[ChainId]*lens.ChainClient
type ChainInitializer = func(logger *zap.Logger, config ChainConfig, homePath string) (*lens.ChainClient, error)

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
			logger.Fatal("failed to connect to the chain", zap.Error(err))
			return nil, err
		}

		_, ok := app.clients[chain.ChainId]
		if ok {
			return nil, fmt.Errorf("chain with id %s already exists", chain.ChainId)
		}
		app.clients[chain.ChainId] = cc
	}

	app.Server = NewServer(app.logger, app.config, app.clients)

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

		// check sender is the valid address
		addr, err := sdk.GetFromBech32(chain.Sender, chain.AccountPrefix)
		if err != nil {
			return err
		}

		// check drop coin is the valid coin
		dropCoin, err := sdk.ParseCoinNormalized(chain.DropCoin)
		if err != nil {
			return err
		}

		// check the denom of `dropCoin` is the same with native currency
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

func (a *App) runRPCServer() error {
	grpcAddr := fmt.Sprintf("0.0.0.0:%d", a.config.Server.Grpc.Port)

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

	return nil
}

func (a *App) runHTTPServer(ctx context.Context) error {
	httpAddr := fmt.Sprintf(":%d", a.config.Server.Http.Port)
	mux := runtime.NewServeMux()
	endpoint := fmt.Sprintf("localhost:%d", a.config.Server.Grpc.Port)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := faucetpb.RegisterFaucetServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	}

	// serve frontend static files
	if err = a.serveFrontend(mux); err != nil {
		return err
	}

	// enable cors if `server.allow_cors` is true
	var handler http.Handler
	if a.config.Server.AllowCors {
		handler = cors.Default().Handler(mux)
	} else {
		handler = mux
	}

	httpSv := http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}

	go func() {
		a.logger.Info("start to serve http Server", zap.String("addr", httpAddr))
		if err := httpSv.ListenAndServe(); err != nil {
			a.logger.Fatal("failed to serve HTTP Server", zap.Error(err))
			a.cancelFunc()
		}
	}()

	return nil
}

func (a *App) Run() error {
	err := a.validateChainConditions()
	if err != nil {
		return fmt.Errorf("validation check failed on chains: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel

	if err := a.runRPCServer(); err != nil {
		return err
	}

	if err := a.runHTTPServer(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) serveFrontend(mux *runtime.ServeMux) error {
	if err := mux.HandlePath("GET", "/", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// serve static file
		http.ServeFile(w, r, "frontend/build/index.html")
	}); err != nil {
		return err
	}

	if err := mux.HandlePath("GET", "/static/js/**", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "frontend/build/"+r.URL.Path)
	}); err != nil {
		return err
	}

	if err := mux.HandlePath("GET", "/static/css/**", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		fmt.Printf("path: %s", r.URL.Path)
		http.ServeFile(w, r, "frontend/build/"+r.URL.Path)
	}); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	defer a.cancelFunc()
	a.logger.Info("shutting down the application")
	err := a.logger.Sync()
	if err != nil {
		a.logger.Panic("logging synchronization failed", zap.Error(err))
	}
}
