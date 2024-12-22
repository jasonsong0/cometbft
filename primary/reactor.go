package primary

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/clist"
	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	primaryproto "github.com/cometbft/cometbft/proto/tendermint/primary"
	"github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/worker"
	"golang.org/x/sync/semaphore"
)

const (
	//TODO: make a config. same or bigger than batch timeout?
	primaryTimeout      = 1000 * time.Millisecond
	primaryClearTimeout = 30 * time.Second

	maxPrimaryMsgSize  = 20000000 // 20MB
	triggerPrimarySize = int64(18000000)
)

// Reactor handles batch tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping batch to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config *cfg.MempoolConfig

	CurrentDagRound int64
	parentMtx       cmtsync.RWMutex
	LastParents     map[p2p.ID][]byte // ID(header proposer) -> cert digest
	CurrentParents  map[p2p.ID][]byte

	//
	DagProcMtx sync.Mutex 
	DagHeaderProc map[types.DagHeaderKey]*types.DagHeaderProc

	// notify workerIndex if any batch is ready
	// TODO: make a multiple batchpool for polling multi workers
	batchReadyQueue chan struct{ uint64 }
	batchpool       worker.Batchpool

	myID p2p.ID

	mtx     cmtsync.RWMutex
	peerMap map[p2p.ID]p2p.Peer
}

// NewReactor returns a new Reactor with the given config and mempool.
// TODO: for poc v1, all worker has same p2p.ID. assupmt that all worker run on same node.
func NewReactor(config *cfg.MempoolConfig) *Reactor {
	primaryR := &Reactor{
		config: config,
	}
	primaryR.BaseReactor = *p2p.NewBaseReactor("Primary", primaryR)

	primaryR.LastParents = make(map[p2p.ID][]byte) // have to get from other reactor
	primaryR.CurrentParents = make(map[p2p.ID][]byte)

	primaryR.DagHeaderProc = make(map[types.DagHeaderKey]*types.DagHeaderProc)

	return primaryR
}

// InitPeer implements Reactor by creating a state for the peer.
func (primR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	primR.mtx.Lock()
	primR.peerMap[peer.ID()] = peer
	primR.mtx.Unlock()
	return peer
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (primR *Reactor) SetLogger(l log.Logger) {
	primR.Logger = l
}

func (primR *Reactor) PrimaryRoutine() {

	myID := primR.Switch.NodeInfo().ID()

	ticker := time.NewTicker(primaryTimeout) // timeout to make a new header
	cleanupTicker := time.NewTicker(primaryClearTimeout)
	// condition: timeout or maxByte or maxGas
	for {
		select {
		case <-ticker.C:

				//make a empty header
				newHeader := primaryproto.PrimaryDagHeader{
					DagRound:      primR.CurrentDagRound,
					AuthorAddress: []byte(primR.myID),
				}
				p := make([][]byte, 0)
				primR.parentMtx.RLock()
				for _, v := range primR.CurrentParents {
					p = append(p, v)
				}
				primR.parentMtx.RUnlock()
				newHeader.Parents = p

				batchDigests := make([][]byte) 
				// TODO: config maxGas
				for i, b := range primR.batchpool.ReapMaxBytesMaxGas(maxPrimaryMsgSize, 100000000) {
					if b.
				}

				//make a dag proc
				newProc := types.DagHeaderProc{
					Hdr: newHeader,
					Votes: make(map[string]primaryproto.DagVote),
					VoteWeight: make(map[string]float64),
					PrimaryCreated: time.Now(),
				}


				workerR.Logger.Info("add batch by timeout", "batchKey", newBatch.batch.BatchKey)
			}

		case <-primR.batchpool.BatchesAvailable():
			// check front batch set acked.
			// should clear the tx notification for getting new txs from mempool

			if workerR.mempool.SizeBytes() >= triggerBatchSize {
				newBatch := workerR.mempool.makeBatchpoolBatch(owner)

				if newBatch == nil {
					continue
				}

				workerR.ackWaitingMtx.Lock()
				workerR.ackWaitingBatch[newBatch.batch.BatchKey] = newBatch
				workerR.ackWaitingMtx.Unlock()

				workerR.batchpool.Lock()
				workerR.batchpool.addBatch(newBatch)
				workerR.batchpool.Unlock()

			}

			workerR.mempool.Update(0, nil, nil, nil, nil)

		case <-cleanupTicker.C:
			//TODO: temp logic for non-acked batch. need to add any logic for dropped batch?
			// clean up old batches
			// if batch is not acked within 5min, remove it from ackWaitingBatch
			for key, batch := range workerR.ackWaitingBatch {
				if time.Since(batch.batch.BatchCreated) > 5*time.Minute {
					workerR.ackWaitingMtx.Lock()
					workerR.ackWaitingBatch[key] = nil
					workerR.ackWaitingMtx.Unlock()
					workerR.Logger.Info("remove old batch", "batchKey", key)
				}
			}

		}

	}
}

// OnStart implements p2p.BaseReactor.
func (primR *Reactor) OnStart() error {

	primR.myID = primR.Switch.NodeInfo().ID()

	ticker := time.NewTicker(primaryTimeout)
	//check to add batch to batchpool
	// condition: timeout or maxByte or maxGas
	for {
		select {
		case <-ticker.C:

			//make a empty header
			newHeader := primaryproto.PrimaryDagHeader{
				DagRound:      primR.CurrentDagRound,
				AuthorAddress: []byte(primR.myID),
			}
			p := make([][]byte, 0)
			primR.parentMtx.RLock()
			for _, v := range primR.CurrentParents {
				p = append(p, v)
			}
			primR.parentMtx.RUnlock()
			newHeader.Parents = p

			newBatch := primR.mempool.makeBatchpoolBatch(owner)

			if newBatch == nil {
				continue
			}

			workerR.batchpool.Lock()
			workerR.batchpool.addBatch(newBatch)
			workerR.batchpool.Unlock()

		case <-workerR.mempool.TxsAvailable():

			if workerR.mempool.SizeBytes() >= triggerBatchSize {
				newBatch := workerR.mempool.makeBatchpoolBatch(owner)

				if newBatch == nil {
					continue
				}

				workerR.batchpool.Lock()
				workerR.batchpool.addBatch(newBatch)
				workerR.batchpool.Unlock()

			}

		}

	}

	return nil
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
func (workerR *Reactor) GetChannels() []*p2p.ChannelDescriptor {

	return []*p2p.ChannelDescriptor{
		{
			ID:                  BatchpoolChannelBase + workerR.workerIndex,
			Priority:            5,
			RecvMessageCapacity: maxBatchMsgSize,
			MessageType:         &protoworker.Message{},
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
	workerR.idxs.Reclaim(peer)
	// broadcast routine checks if peer is gone and returns
}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (workerR *Reactor) Receive(e p2p.Envelope) {
	workerR.Logger.Debug("Receive", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)

	p2pID := workerR.Switch.NodeInfo().ID()

	switch msg := e.Message.(type) {
	case *protoworker.BatchData:

		va := e.VoteA.ToProto()
		vb := e.VoteB.ToProto()
		// Signatures must be valid
		if !pubKey.VerifySignature(types.VoteSignBytes(chainID, va), e.VoteA.Signature) {
			return fmt.Errorf("verifying VoteA: %w", types.ErrVoteInvalidSignature)
		}
		if !pubKey.VerifySignature(types.VoteSignBytes(chainID, vb), e.VoteB.Signature) {
			return fmt.Errorf("verifying VoteB: %w", types.ErrVoteInvalidSignature)
		}

		var err error

		recvTime := time.Now()
		protoTxs := msg.GetTxs()
		if len(protoTxs) == 0 {
			workerR.Logger.Error("received empty txs from peer", "src", e.Src)
			return
		}
		txInfo := TxInfo{SenderID: workerR.idxs.GetIdxForPeer(e.Src)}
		if e.Src != nil {
			txInfo.SenderP2PID = e.Src.ID()
		}

		// check owner, ownerID, workerIdx
		owner := msg.GetBatchOwner()
		ownerArr := strings.Split(owner, "|")
		if len(ownerArr) != 2 {
			workerR.Logger.Error("received invalid owner", "owner", owner)
			return
		}
		ownerID := p2p.ID(ownerArr[0])
		ownerIndex, err := strconv.Atoi(ownerArr[1])
		if ownerID != e.Src.ID() || err != nil || byte(ownerIndex) != workerR.workerIndex {
			workerR.Logger.Error("received invalid owner", "owner", owner, "src", e.Src, "workerIndex", workerR.workerIndex, "err", err)
			return
		}

		//TODO: need to check all txs in batch
		for _, tx := range protoTxs {
			ntx := types.Tx(tx)
			//TODO: check tx in parallel. might be use goroutine
			err = workerR.mempool.CheckTx(ntx, nil, txInfo)
			if errors.Is(err, ErrTxInCache) {
				workerR.Logger.Debug("Tx already exists in cache", "tx", ntx.String())
			} else if err != nil {
				workerR.Logger.Info("Could not check tx", "tx", ntx.String(), "err", err)
			}
		}

		//send ack to batch
		//TODO: get checktx result asynchoronously. and send ack
		myId := string(p2pID) + "|" + string(workerR.workerIndex)
		ack := &protoworker.BatchAck{
			BatchOwner:  myId,
			BatchKey:    msg.GetBatchKey(),
			BatchRecved: recvTime,
		}

		success := e.Src.Send(p2p.Envelope{
			ChannelID: BatchpoolChannelBase + workerR.workerIndex,
			Message:   ack,
		})
		if !success {
			//TODO: retry
			workerR.Logger.Error("failed to send ack", "src", e.Src)
		}

	case *protoworker.BatchAck:

		//TODO: check recvTime and ackTime. RTT can be used to
		recvTime := time.Now()

		// check owner, ownerID, workerIdx
		owner := msg.GetBatchOwner()
		ownerArr := strings.Split(owner, "|")
		if len(ownerArr) != 2 {
			workerR.Logger.Error("received invalid owner", "owner", owner)
			return
		}
		ownerID := p2p.ID(ownerArr[0])
		ownerIndex, err := strconv.Atoi(ownerArr[1])
		if ownerID != e.Src.ID() || err != nil || byte(ownerIndex) != workerR.workerIndex {
			workerR.Logger.Error("received invalid owner", "owner", owner, "src", e.Src, "workerIndex", workerR.workerIndex, "err", err)
			return
		}

		//check valid batchKey in batchpool
		//deleted or already move to primary
		var bk types.BatchKey
		batchKey := msg.GetBatchKey()
		copy(bk[:], batchKey)
		srcIdx := workerR.idxs.GetIdxForPeerByID(e.Src.ID())

		workerR.batchpool.Lock()
		batchEL, exist := workerR.batchpool.getCElement(bk)
		if !exist {
			workerR.batchpool.Unlock()
			workerR.Logger.Error("received invalid batchKey", "batchKey", batchKey)
			return
		}
		batch := batchEL.Value.(*batchpoolBatch)

		batch.acked[srcIdx] = 1.0 //TODO: change to voting power
		if batch.isMaj23() {
			b := batch.batch
			b.BatchConfirmed = recvTime

			workerR.batchpool.Update(batch)
		}

		workerR.batchpool.Unlock()

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

// Send new batchpool batch to peer.
func (workerR *Reactor) broadcastTxRoutine(peer p2p.Peer) {
	peerIdx := workerR.idxs.GetIdxForPeer(peer)

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

		if !b.isSender(peerIdx) {
			success := peer.Send(p2p.Envelope{
				ChannelID: BatchpoolChannelBase + workerR.workerIndex,
				Message:   &protoworker.Batch{Batch: [][]byte{b.batch}},
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
