package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/clist"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	protomem "github.com/cometbft/cometbft/proto/tendermint/mempool"
	"github.com/cometbft/cometbft/types"
	"golang.org/x/sync/semaphore"
)

// Reactor handles batch tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping batch to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config    *cfg.MempoolConfig
	batchpool *CListBatchpool
	mempool   *CListMempool
	ids       *workerIDs

	// Semaphores to keep track of how many connections to peers are active for broadcasting
	// transactions. Each semaphore has a capacity that puts an upper bound on the number of
	// connections for different groups of peers.
	activePersistentPeersSemaphore    *semaphore.Weighted
	activeNonPersistentPeersSemaphore *semaphore.Weighted
}

// NewReactor returns a new Reactor with the given config and mempool.
// TODO: for poc v1, all worker has same p2p.ID. assupmt that all worker run on same node.
func NewReactor(config *cfg.MempoolConfig, batchpool *CListBatchpool, mempool *CListMempool) *Reactor {
	workerR := &Reactor{
		config:    config,
		batchpool: batchpool,
		mempool:   mempool,
		ids:       newWorkerIDs(),
	}
	workerR.BaseReactor = *p2p.NewBaseReactor("Worker", workerR)
	workerR.activePersistentPeersSemaphore = semaphore.NewWeighted(int64(workerR.config.ExperimentalMaxGossipConnectionsToPersistentPeers))
	workerR.activeNonPersistentPeersSemaphore = semaphore.NewWeighted(int64(workerR.config.ExperimentalMaxGossipConnectionsToNonPersistentPeers))

	return workerR
}

// InitPeer implements Reactor by creating a state for the peer.
func (workerR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	workerR.ids.ReserveForPeer(peer)
	return peer
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (workerR *Reactor) SetLogger(l log.Logger) {
	workerR.Logger = l
	workerR.mempool.SetLogger(l)
}

// OnStart implements p2p.BaseReactor.
func (workerR *Reactor) OnStart() error {
	if !workerR.config.Broadcast {
		workerR.Logger.Info("Tx broadcasting is disabled")
	}
	return nil
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
func (workerR *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	largestTx := make([]byte, workerR.config.MaxTxBytes)
	batchMsg := protomem.Message{
		Sum: &protomem.Message_Txs{
			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
		},
	}

	return []*p2p.ChannelDescriptor{
		{
			ID:                  BatchpoolChannel,
			Priority:            5,
			RecvMessageCapacity: batchMsg.Size(),
			MessageType:         &protomem.Message{},
		},
	}
}

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all txs are forwarded to the given peer.
func (workerR *Reactor) AddPeer(peer p2p.Peer) {
	if workerR.config.Broadcast {
		go func() {
			// Always forward transactions to unconditional peers.
			if !workerR.Switch.IsPeerUnconditional(peer.ID()) {
				// Depending on the type of peer, we choose a semaphore to limit the gossiping peers.
				var peerSemaphore *semaphore.Weighted
				if peer.IsPersistent() && workerR.config.ExperimentalMaxGossipConnectionsToPersistentPeers > 0 {
					peerSemaphore = workerR.activePersistentPeersSemaphore
				} else if !peer.IsPersistent() && workerR.config.ExperimentalMaxGossipConnectionsToNonPersistentPeers > 0 {
					peerSemaphore = workerR.activeNonPersistentPeersSemaphore
				}

				if peerSemaphore != nil {
					for peer.IsRunning() {
						// Block on the semaphore until a slot is available to start gossiping with this peer.
						// Do not block indefinitely, in case the peer is disconnected before gossiping starts.
						ctxTimeout, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
						// Block sending transactions to peer until one of the connections become
						// available in the semaphore.
						err := peerSemaphore.Acquire(ctxTimeout, 1)
						cancel()

						if err != nil {
							continue
						}

						// Release semaphore to allow other peer to start sending transactions.
						defer peerSemaphore.Release(1)
						break
					}
				}
			}

			workerR.mempool.metrics.ActiveOutboundConnections.Add(1)
			defer workerR.mempool.metrics.ActiveOutboundConnections.Add(-1)
			workerR.broadcastTxRoutine(peer)
		}()
	}
}

// RemovePeer implements Reactor.
func (workerR *Reactor) RemovePeer(peer p2p.Peer, _ interface{}) {
	workerR.ids.Reclaim(peer)
	// broadcast routine checks if peer is gone and returns
}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (workerR *Reactor) Receive(e p2p.Envelope) {
	workerR.Logger.Debug("Receive", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
	switch msg := e.Message.(type) {
	case *protomem.Txs:
		protoTxs := msg.GetTxs()
		if len(protoTxs) == 0 {
			workerR.Logger.Error("received empty txs from peer", "src", e.Src)
			return
		}
		txInfo := TxInfo{SenderID: workerR.ids.GetForPeer(e.Src)}
		if e.Src != nil {
			txInfo.SenderP2PID = e.Src.ID()
		}

		var err error
		for _, tx := range protoTxs {
			ntx := types.Tx(tx)
			err = workerR.mempool.CheckTx(ntx, nil, txInfo)
			if errors.Is(err, ErrTxInCache) {
				workerR.Logger.Debug("Tx already exists in cache", "tx", ntx.String())
			} else if err != nil {
				workerR.Logger.Info("Could not check tx", "tx", ntx.String(), "err", err)
			}
		}
	default:
		workerR.Logger.Error("unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		workerR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
		return
	}

	// broadcasting happens from go routines per peer
}

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// Send new mempool txs to peer.
func (workerR *Reactor) broadcastTxRoutine(peer p2p.Peer) {
	peerID := workerR.ids.GetForPeer(peer)
	var next *clist.CElement

	for {
		// In case of both next.NextWaitChan() and peer.Quit() are variable at the same time
		if !workerR.IsRunning() || !peer.IsRunning() {
			return
		}

		// This happens because the CElement we were looking at got garbage
		// collected (removed). That is, .NextWait() returned nil. Go ahead and
		// start from the beginning.
		if next == nil {
			select {
			case <-workerR.batchpool.BatchesWaitChan(): // Wait until a tx is available
				if next = workerR.batchpool.BatchesFront(); next == nil {
					continue
				}
			case <-peer.Quit():
				return
			case <-workerR.Quit():
				return
			}
		}

		// Make sure the peer is up to date.
		peerState, ok := peer.Get(types.PeerStateKey).(PeerState)
		if !ok {
			// Peer does not have a state yet. We set it in the consensus reactor, but
			// when we add peer in Switch, the order we call reactors#AddPeer is
			// different every time due to us using a map. Sometimes other reactors
			// will be initialized before the consensus reactor. We should wait a few
			// milliseconds and retry.
			time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		// Allow for a lag of 1 block.
		b := next.Value.(*batchpoolBatch)
		if peerState.GetHeight() < b.Height()-1 {
			time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		// NOTE: Transaction batching was disabled due to
		// https://github.com/tendermint/tendermint/issues/5796

		if !b.isSender(peerID) {
			success := peer.Send(p2p.Envelope{
				ChannelID: BatchpoolChannel,
				Message:   &protomem.Batches{Batches: [][]byte{b.batch}},
			})
			if !success {
				time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
				continue
			}
		}

		select {
		case <-next.NextWaitChan():
			// see the start of the for loop for nil check
			next = next.Next()
		case <-peer.Quit():
			return
		case <-workerR.Quit():
			return
		}
	}
}

// TxsMessage is a Message containing transactions.
type TxsMessage struct {
	Txs []types.Tx
}

// String returns a string representation of the TxsMessage.
func (m *TxsMessage) String() string {
	return fmt.Sprintf("[TxsMessage %v]", m.Txs)
}
