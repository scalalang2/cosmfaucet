package core

import (
	"context"
	"os"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

// RPCClient is an interface for the RPC client
// It's based on lens client pacakge
type RPCClient interface {
	SendMsg(ctx context.Context, msg sdk.Msg) (*sdk.TxResponse, error)
	QueryBalanceWithDenomTraces(ctx context.Context, address sdk.AccAddress, pageReq *query.PageRequest) (sdk.Coins, error)
	MustEncodeAccAddr(addr sdk.AccAddress) string
}

// chain initializer
var newChainClient = newLensClient

// newLensClient create a client for the cosmos blockchain
func newLensClient(logger *zap.Logger, config ChainConfig, homePath string) (RPCClient, error) {
	if !strings.HasPrefix(config.RpcEndpoint, "http") {
		return nil, errInvalidEndpoint{rpc: config.RpcEndpoint}
	}

	cfg := lens.ChainClientConfig{
		Key:            config.KeyName,
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

	// ignore the error
	addr, _ := cc.RestoreKey(config.KeyName, config.Key, 118)
	logger.Info("master wallet is restored", zap.String("address", addr))

	return cc, nil
}
