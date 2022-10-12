package core

import (
	"context"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	mux sync.Mutex
	log *zap.Logger

	faucetpb.FaucetServiceServer
	config  *RootConfig
	clients ChainClients
}

func NewServer(log *zap.Logger, config *RootConfig, clients ChainClients) *Server {
	return &Server{
		log:     log,
		config:  config,
		clients: clients,
	}
}

// GiveMe sends a `BankMsg` transaction to the chain to send some tokens to the given address
// It blocks the request if the user is given the token in the last 24 hours.
func (s *Server) GiveMe(ctx context.Context, request *faucetpb.GiveMeRequest) (*faucetpb.GiveMeResponse, error) {
	client, ok := s.clients[request.ChainId]
	if !ok {
		return nil, status.Error(codes.NotFound, "chain not supported")
	}

	// find config from RootConfig
	var chainConfig *ChainConfig
	for _, chain := range s.config.Chains {
		if chain.ChainId == request.ChainId {
			chainConfig = &chain
			break
		}
	}

	// validate address format
	acc, err := sdk.GetFromBech32(request.Address, chainConfig.AccountPrefix)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	// TODO: check if the user is already given the token in the last 24 hours
	// send the bank msg transaction
	s.mux.Lock()
	defer s.mux.Unlock()

	from, err := sdk.GetFromBech32(chainConfig.Sender, chainConfig.AccountPrefix)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid sender address | this is unexpected error, please inform to the admin.")
	}

	coin, err := sdk.ParseCoinNormalized(chainConfig.DropCoin)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid coin format | this is unexpected error, please inform to the admin.")
	}

	msg := banktypes.NewMsgSend(from, acc, []sdk.Coin{coin})
	txResponse, err := client.SendMsg(ctx, msg)
	if err != nil {
		s.log.Error("failed to send transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to send transaction, please try later")
	}

	s.log.Info("BankMsg transaction has been executed",
		zap.String("tx_hash", txResponse.TxHash),
		zap.String("to_address", request.Address),
		zap.String("chain", request.ChainId),
	)

	return &faucetpb.GiveMeResponse{}, nil
}

// Chains returns all supported chains
func (s *Server) Chains(ctx context.Context, request *faucetpb.GetChainsRequest) (*faucetpb.GetChainsResponse, error) {
	res := make([]*faucetpb.Chain, 0)
	for _, chain := range s.config.Chains {
		res = append(res, &faucetpb.Chain{
			Name:    chain.Name,
			ChainId: chain.ChainId,
		})
	}

	return &faucetpb.GetChainsResponse{
		Chains: res,
	}, nil
}
