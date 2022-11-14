package core

import (
	"os"
	"strings"

	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

// chain initializer
var newChainClient = newLensClient

// newLensClient create a client for the cosmos blockchain
func newLensClient(logger *zap.Logger, config ChainConfig, homePath string) (*lens.ChainClient, error) {
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
