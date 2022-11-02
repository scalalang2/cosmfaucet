package core

import (
	"context"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	mux     sync.Mutex
	log     *zap.Logger
	limiter *Limiter

	faucetpb.FaucetServiceServer
	config  *RootConfig
	clients ChainClients
}

func NewServer(log *zap.Logger, config *RootConfig, clients ChainClients) *Server {
	chains := make([]ChainId, 0)
	for _, chainCfg := range config.Chains {
		chains = append(chains, chainCfg.ChainId)
	}

	var limiter *Limiter
	if config.Server.Limit.Enabled {
		limiter = NewLimiter(chains, config.Server.Limit.Period)
	}

	return &Server{
		log:     log,
		limiter: limiter,
		config:  config,
		clients: clients,
	}
}

// GiveMe sends a `BankMsg` transaction to the chain to send some tokens to the given address
// It blocks the request if the user is given the token in the last 24 hours.
func (s *Server) GiveMe(ctx context.Context, request *faucetpb.GiveMeRequest) (*faucetpb.GiveMeResponse, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get peer from context")
	}

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

	from, err := sdk.GetFromBech32(chainConfig.Sender, chainConfig.AccountPrefix)
	if err != nil {
		s.log.Error("invalid sender address", zap.Error(err))
		return nil, status.Error(codes.Internal, "invalid sender address | this is unexpected error, please inform to the admin.")
	}

	coin, err := sdk.ParseCoinNormalized(chainConfig.DropCoin)
	if err != nil {
		s.log.Error("invalid coin format", zap.Error(err))
		return nil, status.Error(codes.Internal, "invalid coin format | this is unexpected error, please inform to the admin.")
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if s.limiter != nil {
		remoteAddr := p.Addr.String()
		if !s.limiter.IsAllowed(request.ChainId, remoteAddr) {
			return nil, status.Error(codes.PermissionDenied, "user cannot request token more than once during specific period of time")
		}
	}

	s.log.Info("trying to send tokens",
		zap.String("from", client.MustEncodeAccAddr(from)),
		zap.String("to", client.MustEncodeAccAddr(acc)),
		zap.String("coin", coin.String()))

	msg := &banktypes.MsgSend{
		FromAddress: client.MustEncodeAccAddr(from),
		ToAddress:   client.MustEncodeAccAddr(acc),
		Amount:      []sdk.Coin{coin},
	}

	txResponse, err := client.SendMsg(ctx, msg)
	if err != nil {
		s.log.Error("failed to send transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to send transaction, please try later")
	}

	if s.limiter != nil {
		remoteAddr := p.Addr.String()
		s.limiter.AddRequest(request.ChainId, remoteAddr)
	}

	s.log.Info("BankMsg transaction has been executed",
		zap.String("tx_hash", txResponse.TxHash),
		zap.String("to_address", request.Address),
		zap.String("chain", request.ChainId),
	)

	return &faucetpb.GiveMeResponse{
		TxHash: txResponse.TxHash,
	}, nil
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
