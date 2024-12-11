package opsimulator

import (
	"context"
	"fmt"

	"github.com/asaskevich/EventBus"
	"github.com/ethereum-optimism/optimism/op-service/tasks"
	"github.com/ethereum-optimism/supersim/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type L1ToL2MessageIndexer struct {
	log          log.Logger
	storeManager *L1DepositStoreManager
	eb           EventBus.Bus
	l1Chain      config.Chain
	tasks        tasks.Group
	tasksCtx     context.Context
	tasksCancel  context.CancelFunc
}

func NewL1ToL2MessageIndexer(log log.Logger, l1Chain config.Chain) *L1ToL2MessageIndexer {
	tasksCtx, tasksCancel := context.WithCancel(context.Background())

	return &L1ToL2MessageIndexer{
		log:          log,
		storeManager: NewL1DepositStoreManager(),
		eb:           EventBus.New(),
		l1Chain:      l1Chain,
		tasks: tasks.Group{
			HandleCrit: func(err error) {
				fmt.Printf("unhandled indexer error: %v\n", err)
			},
		},
		tasksCtx:    tasksCtx,
		tasksCancel: tasksCancel,
	}
}

func (i *L1ToL2MessageIndexer) Start(ctx context.Context) error {

	i.tasks.Go(func() error {
		depositTxCh := make(chan *types.DepositTx)
		l1DepositTxnLogCh := make(chan *types.Log)
		portalAddress := common.Address(i.l1Chain.Config().L2Config.L1Addresses.OptimismPortalProxy)
		sub, err := SubscribeDepositTx(i.tasksCtx, i.l1Chain.EthClient(), portalAddress, depositTxCh, l1DepositTxnLogCh)

		if err != nil {
			return fmt.Errorf("failed to subscribe to deposit tx: %w", err)
		}

		chainID := i.l1Chain.Config().ChainID

		for {
			select {
			case dep := <-depositTxCh:

				if err := i.processEvent(dep, chainID); err != nil {
					fmt.Printf("failed to process log: %v\n", err)
				}

			case <-i.tasksCtx.Done():
				sub.Unsubscribe()
			}
		}
	})

	return nil
}

func (i *L1ToL2MessageIndexer) Stop(ctx context.Context) error {
	i.tasksCancel()
	return nil
}

func depositMessageInfoKey() string {
	return fmt.Sprintln("DepositMessageKey")
}

func (i *L1ToL2MessageIndexer) SubscribeDepositMessage(depositMessageChan chan<- *types.DepositTx) (func(), error) {
	return i.createSubscription(depositMessageInfoKey(), depositMessageChan)
}

func (i *L1ToL2MessageIndexer) createSubscription(key string, depositMessageChan chan<- *types.DepositTx) (func(), error) {
	handler := func(e *types.DepositTx) {
		depositMessageChan <- e
	}

	if err := i.eb.Subscribe(key, handler); err != nil {
		return nil, fmt.Errorf("failed to create subscription %s: %w", key, err)
	}

	return func() {
		_ = i.eb.Unsubscribe(key, handler)
	}, nil
}

func (i *L1ToL2MessageIndexer) processEvent(dep *types.DepositTx, chainID uint64) error {

	depTx := types.NewTx(dep)
	i.log.Debug("observed deposit event on L1", "hash", depTx.Hash().String())

	if err := i.storeManager.Set(depTx.Hash(), dep); err != nil {
		i.log.Error("failed to store deposit tx to chain: %w", "chain.id", chainID, "err", err)
		return err
	}

	i.eb.Publish(depositMessageInfoKey(), depTx)

	return nil
}
