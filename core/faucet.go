package core

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
)

type work struct {
	chainId ChainId
	detail  workDetail
}

type workDetail interface {
	msg() sdk.Msg
}

type transferWork struct {
	fromAddress string
	toAddress   string
	amount      []sdk.Coin
}

func (w *transferWork) msg() sdk.Msg {
	return &banktypes.MsgSend{
		FromAddress: w.fromAddress,
		ToAddress:   w.toAddress,
		Amount:      w.amount,
	}
}

type Faucet struct {
	logger  *zap.Logger
	clients ChainClients
	works   map[ChainId]chan *work
}

func NewFaucet(logger *zap.Logger, clients ChainClients, buf int) *Faucet {
	works := make(map[ChainId]chan *work)
	for chainId := range clients {
		works[chainId] = make(chan *work, buf)
	}

	return &Faucet{
		logger:  logger,
		clients: clients,
		works:   works,
	}
}

func (f *Faucet) run() {
	for chainId, workCh := range f.works {
		go f.runWorker(chainId, workCh)
	}
}

func (f *Faucet) runWorker(chainId ChainId, workCh chan *work) {
	tick := time.NewTicker(time.Second)
	messages := make([]sdk.Msg, 0)

	for {
		select {
		case <-tick.C:
			if len(messages) > 0 {
				tx, err := f.clients[chainId].SendMsgs(context.Background(), messages)
				if err != nil {
					f.logger.Error("failed to send transaction, failed messages will be removed from queue",
						zap.String("chain_id", string(chainId)),
						zap.Int("messages", len(messages)),
						zap.Error(err))
				} else {
					f.logger.Info("sent transaction",
						zap.String("chain_id", string(chainId)),
						zap.String("tx_hash", tx.TxHash),
						zap.Int("messages", len(messages)))
				}

				messages = make([]sdk.Msg, 0)
			}
		case w := <-workCh:
			messages = append(messages, w.detail.msg())
		}
	}
}

func (f *Faucet) sendTask(chainId ChainId, work *work) {
	f.works[chainId] <- work
}
