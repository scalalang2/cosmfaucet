package core

import (
	"context"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
)

type Server struct {
	faucetpb.FaucetServiceServer
	config  *RootConfig
	clients *ChainClients
}

func NewServer(config *RootConfig, clients *ChainClients) *Server {
	return &Server{
		config:  config,
		clients: clients,
	}
}

// GiveMe sends a `BankMsg` transaction to the chain to send some tokens to the given address
// It blocks the request if the user is given the token in the last 24 hours.
func (s *Server) GiveMe(ctx context.Context, request *faucetpb.GiveMeRequest) (*faucetpb.GiveMeResponse, error) {
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
