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

func (s *Server) GiveMe(ctx context.Context, request *faucetpb.GiveMeRequest) (*faucetpb.GiveMeResponse, error) {
	return &faucetpb.GiveMeResponse{}, nil
}

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
