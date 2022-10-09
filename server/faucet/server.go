package faucet

import (
	"context"
	"github.com/scalalang2/cosmfaucet/gen/proto/faucetpb"
)

type Server struct {
	faucetpb.FaucetServiceServer
}

func New() *Server {
	return &Server{}
}

func (s Server) GiveMe(ctx context.Context, request *faucetpb.GiveMeRequest) (*faucetpb.GiveMeResponse, error) {
	return &faucetpb.GiveMeResponse{}, nil
}

func (s Server) Chains(ctx context.Context, request *faucetpb.GetChainsRequest) (*faucetpb.GetChainsResponse, error) {
	return &faucetpb.GetChainsResponse{
		Chains: []*faucetpb.Chain{
			{
				Name:    "test",
				ChainId: "testchain",
			},
		},
	}, nil
}
