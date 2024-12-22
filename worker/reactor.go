package worker

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/clist"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	protoworker "github.com/cometbft/cometbft/proto/tendermint/worker"
	"github.com/cometbft/cometbft/types"
)

const (
	//TODO: make a config
	batchTimeout      = 1000 * time.Millisecond
	batchClearTimeout = 300 * time.Second

	maxBatchMsgSize = 10000000 // 10MB

	triggerBatchSize     = int64(90000)
	maxBytesOfTxsInBatch = 100000
	maxGasOfTxsInBatch   = 100000
)

// Reactor handles batch tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping batch to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config      *cfg.MempoolConfig
	batchpool   *CListBatchpool
	mempool     *CListMempool
	idxs        *workerIDXs
	workerIndex byte // used for channel id

	ackWaitingBatch map[types.BatchKey]*batchpoolBatch
	ackWaitingMtx   sync.Mutex

	batchReadyNotifier chan struct{}
}

// NewReactor returns a new Reactor with the given config and mempool.
// TODO: for poc v1, all worker has same p2p.ID. assupmt that all worker run on same node.
func NewReactor(config *cfg.MempoolConfig, batchpool *CListBatchpool, mempool *CListMempool, workerIndex byte) *Reactor {
	workerR := &Reactor{
		config:      config,
		batchpool:   batchpool,
		mempool:     mempool,
		idxs:        newWorkerIDXs(),
		workerIndex: workerIndex,
	}
	workerR.ackWaitingBatch = make(map[types.BatchKey]*batchpoolBatch)
	indexStr := strconv.Itoa(int(workerIndex))
	workerR.BaseReactor = *p2p.NewBaseReactor("Worker"+indexStr, workerR)

	return workerR
}

// InitPeer implements Reactor by creating a state for the peer.
func (workerR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	workerR.idxs.ReserveForPeer(peer)
	return peer
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (workerR *Reactor) SetLogger(l log.Logger) {
	workerR.Logger = l
	workerR.mempool.SetLogger(l)
}

func (workerR *Reactor) WorkerRoutine() {

	myID := workerR.Switch.NodeInfo().ID()
	owner := string(myID) + "|" + string(workerR.workerIndex)

	ticker := time.NewTicker(batchTimeout)
	cleanupTicker := time.NewTicker(batchClearTimeout)
	//check to add batch to batchpool
	// condition: timeout or maxByte or maxGas
	for {
		select {
		case <-ticker.C:

			if workerR.mempool.SizeBytes() != 0 {

				newBatch := workerR.mempool.makeBatchpoolBatch(owner)

				if newBatch == nil {
					continue
				}

				workerR.ackWaitingMtx.Lock()
				workerR.ackWaitingBatch[newBatch.batch.BatchKey] = newBatch
				workerR.ackWaitingMtx.Unlock()

				// add batch to batchpool even if it is not acked.
				// to ensure the batch incoming sequence
				workerR.batchpool.Lock()
				workerR.batchpool.addBatch(newBatch)
				workerR.batchpool.Unlock()

				workerR.Logger.Info("add batch by timeout", "batchKey", newBatch.batch.BatchKey)
			}

		case <-workerR.mempool.TxsAvailable():
			// check condition and make a batch if possible
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
func (workerR *Reactor) OnStart() error {

	go workerR.WorkerRoutine()

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

	// run loop to broadcast batch to peer
	workerR.broadcastTxRoutine(peer)
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
		workerR.Logger.Debug("received batch", "src", e.Src, "batch", msg.BatchKey, "txs", len(protoTxs))
		checkTxStart := time.Now()
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
		checkTxEnd := time.Now()
		checkTxTime := checkTxEnd.Sub(checkTxStart)
		workerR.Logger.Debug("checkTxTime", "time", checkTxTime)

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

		workerR.ackWaitingMtx.Lock()
		waitingBatch, exist := workerR.ackWaitingBatch[bk]
		if !exist {
			workerR.Logger.Error("received invalid batchKey", "batchKey", batchKey)
			workerR.ackWaitingMtx.Unlock()
			return
		}

		waitingBatch.acked[srcIdx] = 1.0 //TODO: change to voting power
		if waitingBatch.isMaj23() {
			b := waitingBatch.batch
			b.BatchConfirmed = recvTime

			waitingBatch.maj23 = true
			// noti to primary. if the batch is not front of the batchpool, no need to notify

			workerR.ackWaitingBatch[bk] = nil

		}
		workerR.ackWaitingMtx.Unlock()

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
			workerR.Logger.Info("broadcastTxRoutine quit", "worker", workerR.workerIndex, "peer", peer.ID())
			return
		}

		// This happens because the CElement we were looking at got garbage
		// collected (removed). That is, .NextWait() returned nil. Go ahead and
		// start from the beginning.
		if next == nil {
			select {
			case <-workerR.batchpool.BatchesWaitChan(): // Wait until a batch is available
				if next = workerR.batchpool.BatchesFront(); next == nil {
					continue
				}
			case <-peer.Quit():
				return
			case <-workerR.Quit():
				return
			}
		}

		// Allow for a lag of 1 block.
		b := next.Value.(*batchpoolBatch)

		// NOTE: Transaction batching was disabled due to
		// https://github.com/tendermint/tendermint/issues/5796
		// TODO: check the issue can be blocker for us
		if !b.isSender(peerIdx) {
			success := peer.Send(p2p.Envelope{
				ChannelID: BatchpoolChannelBase + workerR.workerIndex,
				Message: &protoworker.BatchData{
					Txs:          b.batch.TxArr.ToSliceOfBytes(),
					BatchKey:     b.batch.BatchKey[:],
					BatchOwner:   b.batch.BatchOwner,
					BatchCreated: b.batch.BatchCreated,
				},
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
