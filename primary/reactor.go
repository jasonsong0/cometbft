package primary

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	primaryproto "github.com/cometbft/cometbft/proto/tendermint/primary"
	"github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/worker"
)

const (
	primaryChannel = byte(0x31)

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

	//NOTE: using for sign header. need to change to use Switch if possible
	nodeKey *p2p.NodeKey

	roundMtx                cmtsync.RWMutex
	CurrentDagRound         int64
	LastHeaderDagRound      int64
	LastParents             map[p2p.ID]types.DagCertKey // ID(header proposer) -> cert digests(last dag-round)
	CurrentParents          map[p2p.ID]types.DagCertKey // digests of cert received this dag-round
	CurrentCertSum          big.Float                   // for majority check
	NextRoundToLeaderChange int64
	NextLeader              p2p.ID

	CertCache map[types.DagCertKey]primaryproto.PrimaryDagCert

	//
	DagProcMtx    sync.Mutex
	DagHeaderProc map[types.DagHeaderKey]*types.DagHeaderProc

	// notify workerIndex if any batch is ready
	// TODO: make a multiple batchpool for polling multi workers
	batchReadyQueue chan struct{ uint64 }
	batchpool       worker.Batchpool

	myID p2p.ID

	mtx     cmtsync.RWMutex
	peerMap map[p2p.ID]p2p.Peer
	leader  p2p.ID
}

// NewReactor returns a new Reactor with the given config and mempool.
// TODO: for poc v1, all worker has same p2p.ID. assupmt that all worker run on same node.
// TODO: nodekey from Switch is private. so get an argument now for temporary.
func NewReactor(config *cfg.MempoolConfig, privKey *p2p.NodeKey) *Reactor {
	primaryR := &Reactor{
		config: config,
	}
	primaryR.BaseReactor = *p2p.NewBaseReactor("Primary", primaryR)

	primaryR.LastParents = make(map[p2p.ID]types.DagCertKey) // have to get from other reactor
	primaryR.CurrentParents = make(map[p2p.ID]types.DagCertKey)

	primaryR.DagHeaderProc = make(map[types.DagHeaderKey]*types.DagHeaderProc)

	primaryR.nodeKey = privKey

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

	var lastGenRound, curRound int64

	ticker := time.NewTicker(primaryTimeout) // timeout to make a new header
	//cleanupTicker := time.NewTicker(primaryClearTimeout)
	// condition: timeout or maxByte or maxGas
	for {
		select {
		case <-ticker.C:

			lastGenRound = atomic.LoadInt64(&primR.LastHeaderDagRound)
			curRound = atomic.LoadInt64(&primR.CurrentDagRound)
			if lastGenRound >= curRound {
				primR.Logger.Info("skip header generation", "lastGenRound", lastGenRound, "curRound", curRound)
				continue
			}

			//make a empty header
			newHeader := primaryproto.PrimaryDagHeader{
				DagRound:      curRound,
				AuthorAddress: []byte(primR.myID),
			}
			p := make([][]byte, 0)
			primR.roundMtx.RLock()
			for _, v := range primR.CurrentParents {
				p = append(p, v.ToBytes())
			}
			primR.roundMtx.RUnlock()
			newHeader.Parents = p

			batchDigests := make([][]byte, 0)
			// TODO: config maxGas
			batches, _ := primR.batchpool.ReapMaxBytesMaxGas(maxPrimaryMsgSize, 0)
			for _, b := range batches {
				// batch is already acked and get enough majority votes
				incTime := time.Now()
				primR.Logger.Debug("batch included", "batchKey", b.BatchKey, "latencyE2E", incTime.Sub(b.BatchCreated), "latencyPrimary", incTime.Sub(b.BatchConfirmed))

				batchDigests = append(batchDigests, b.BatchKey[:])
			}
			newHeader.BatchDigests = batchDigests

			//make a dag proc
			newProc := types.DagHeaderProc{
				Hdr:            newHeader,
				Votes:          make(map[string]*primaryproto.DagVote),
				VoteWeight:     make(map[string]float64),
				PrimaryCreated: time.Now(),
			}

			headerKey, err := types.GenDagHeaderKey(&newHeader)
			if err != nil {
				primR.Logger.Error("failed to get header hash", "err", err)
				continue
			}
			primR.DagHeaderProc[headerKey] = &newProc

			primR.Logger.Info("add header by timeout", "headerHash", headerKey, "dagRound", primR.CurrentDagRound, "parents", newHeader.Parents, "batchDigests", batchDigests)

			// broadcast header
			for _, peer := range primR.peerMap {
				peer.Send(p2p.Envelope{
					ChannelID: primaryChannel,
					Message:   &newHeader,
				})
			}

		case <-primR.batchpool.BatchesAvailable():

			lastGenRound = atomic.LoadInt64(&primR.LastHeaderDagRound)
			curRound = atomic.LoadInt64(&primR.CurrentDagRound)
			if lastGenRound >= curRound {
				primR.Logger.Info("skip header generation", "lastGenRound", lastGenRound, "curRound", curRound)
				continue
			}

			// check front batch set acked.
			batches, reapType := primR.batchpool.ReapMaxBytesMaxGas(maxPrimaryMsgSize, 0)
			if reapType != worker.ReapNormal {
				//make a empty header
				newHeader := primaryproto.PrimaryDagHeader{
					DagRound:      primR.CurrentDagRound,
					AuthorAddress: []byte(primR.myID),
				}
				p := make([][]byte, 0)
				primR.roundMtx.RLock()
				for _, v := range primR.CurrentParents {
					p = append(p, v.ToBytes())
				}
				primR.roundMtx.RUnlock()
				newHeader.Parents = p

				batchDigests := make([][]byte, 0)
				for _, b := range batches {
					// batch is already acked and get enough majority votes
					incTime := time.Now()
					primR.Logger.Debug("batch included", "batchKey", b.BatchKey, "latencyE2E", incTime.Sub(b.BatchCreated), "latencyPrimary", incTime.Sub(b.BatchConfirmed))

					batchDigests = append(batchDigests, b.BatchKey[:])
				}
				newHeader.BatchDigests = batchDigests

				//make a dag proc
				newProc := types.DagHeaderProc{
					Hdr:            newHeader,
					Votes:          make(map[string]*primaryproto.DagVote),
					VoteWeight:     make(map[string]float64),
					PrimaryCreated: time.Now(),
				}

				headerKey, err := types.GenDagHeaderKey(&newHeader)
				if err != nil {
					primR.Logger.Error("failed to get header hash", "err", err)
					continue
				}
				primR.DagHeaderProc[headerKey] = &newProc

				primR.Logger.Info("add header by batch full", "headerHash", headerKey, "dagRound", primR.CurrentDagRound, "parents", newHeader.Parents, "batchDigests", batchDigests)

				// broadcast header
				for _, peer := range primR.peerMap {
					peer.Send(p2p.Envelope{
						ChannelID: primaryChannel,
						Message:   &newHeader,
					})
				}

			}
			/*
				case <-cleanupTicker.C:
					//TODO: temp logic for non-acked batch. need to add any logic for dropped batch?
					// clean up old batches
					// if batch is not acked within 5min, remove it from ackWaitingBatch
					for key, batch := range primR.ackWaitingBatch {
						if time.Since(batch.batch.BatchCreated) > 5*time.Minute {
							primR.ackWaitingMtx.Lock()
							primR.ackWaitingBatch[key] = nil
							primR.ackWaitingMtx.Unlock()
							primR.Logger.Info("remove old batch", "batchKey", key)
						}
					}
			*/

		}

	}
}

// OnStart implements p2p.BaseReactor.
func (primR *Reactor) OnStart() error {

	go primR.PrimaryRoutine()

	return nil
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
func (primR *Reactor) GetChannels() []*p2p.ChannelDescriptor {

	return []*p2p.ChannelDescriptor{
		{
			ID:                  primaryChannel,
			Priority:            5,
			RecvMessageCapacity: maxPrimaryMsgSize,
			MessageType:         &primaryproto.Message{},
		},
	}
}

// AddPeer implements Reactor.
// TODO: leader logic might be changed. for poc v1, the leader is fixed with most staked node.
func (primR *Reactor) AddPeer(peer p2p.Peer) {

	// send leader notify at first added
	peer.Send(p2p.Envelope{
		ChannelID: primaryChannel,
		Message:   &primaryproto.PrimaryLeaderNotify{FromDagRound: 1, LeaderAddress: []byte(primR.leader)},
	})

}

// RemovePeer implements Reactor.
func (primR *Reactor) RemovePeer(peer p2p.Peer, _ interface{}) {
	// broadcast routine checks if peer is gone and returns
}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (primR *Reactor) Receive(e p2p.Envelope) {
	primR.Logger.Debug("Receive", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)

	p2pID := primR.Switch.NodeInfo().ID()

	switch msg := e.Message.(type) {
	case *primaryproto.PrimaryDagHeader:

		if len(msg.Parents) == 0 && primR.CurrentDagRound != 0 {
			primR.Logger.Error("received invalid header", "header", msg)
			return
		}

		//TODO: check parents validation
		//for _, p := range msg.Parents {
		//primR.LastParents[e.Src.ID()] = p
		//}

		data, err := msg.Marshal()
		if err != nil {
			primR.Logger.Error("failed to marshal header", "err", err)
			return
		}
		sign, err := primR.nodeKey.PrivKey.Sign(data)
		if err != nil {
			primR.Logger.Error("failed to sign header", "err", err)
			return
		}

		headerHash, err := types.GenDagHeaderKey(msg)
		if err != nil {
			primR.Logger.Error("failed to get header hash", "err", err)
		}
		dagVote := primaryproto.PrimaryDagVote{
			DagHeaderHash: headerHash[:],
			Round:         msg.DagRound,
			OriginAddress: msg.AuthorAddress,
			VoterAddress:  []byte(p2pID),
			Signature:     sign,
		}

		success := e.Src.Send(p2p.Envelope{
			ChannelID: primaryChannel,
			Message:   &dagVote,
		})
		if !success {
			//TODO: retry
			primR.Logger.Error("failed to send vode", "src", e.Src)
		}

	case *primaryproto.PrimaryDagVote:

		if !bytes.Equal(msg.OriginAddress, []byte(e.Src.ID())) {
			primR.Logger.Error("received invalid vote", "vote", msg, "src", e.Src)
			return
		}

		var headerKey types.DagHeaderKey
		copy(headerKey[:], msg.DagHeaderHash[:types.BatchKeySize])

		proc, ok := primR.DagHeaderProc[headerKey]

		if !ok {
			primR.Logger.Error("received invalid vote. not in proc", "vote", msg, "src", e.Src)
			return
		}

		voteOnly := primaryproto.DagVote{
			ValidatorAddress: msg.VoterAddress,
			Signature:        msg.Signature,
		}
		proc.Votes[string(e.Src.ID())] = &voteOnly
		proc.VoteWeight[string(e.Src.ID())] = 1.0 //TODO: change to voting power

		//TODO: check majority votes using voting power
		if len(proc.Votes) > len(primR.peerMap)/2 {

			votes := make([]*primaryproto.DagVote, 0)
			for _, v := range proc.Votes {
				votes = append(votes, v)
			}
			// make a certificate and send to leader
			cert := primaryproto.PrimaryDagCert{
				DagHeaderHash: msg.DagHeaderHash,
				DagHeader:     &proc.Hdr,
				Votes:         votes,
			}

			//TODO: send cert to everyone
			primR.Logger.Info("broadcast certificate", "cert", cert)
			primR.Switch.Broadcast(p2p.Envelope{
				ChannelID: primaryChannel,
				Message:   &cert,
			})

		}

	case *primaryproto.PrimaryDagCert:

		digest, err := types.GenCertDigest(msg)
		if err != nil {
			primR.Logger.Error("failed to get cert digest", "err", err)
			return
		}

		primR.CertCache[types.DagCertKey(digest)] = *msg
		primR.CurrentParents[p2p.ID(msg.DagHeader.AuthorAddress)] = digest
		primR.CurrentCertSum.Add(&primR.CurrentCertSum, new(big.Float).SetInt64(1))

		primR.Logger.Info("add cert", "cert", msg)

		//TODO: use validator voting power
		if primR.CurrentCertSum.Cmp(big.NewFloat(float64(len(primR.peerMap)/2))) > 0 {
			primR.roundMtx.Lock()
			primR.LastParents = primR.CurrentParents
			primR.CurrentParents = make(map[p2p.ID]types.DagCertKey)
			primR.CurrentCertSum.SetInt64(0)
			primR.roundMtx.Unlock()

			//TODO: send to proposer
		}

		//TODO: in original, certification is sent to consensus layer. To check if nessasary to send to consensus layer

	case *primaryproto.PrimaryLeaderNotify:
		//NOTE: for poc v1, leader is fixed with most staked node. this can be used for new peer added.

		nextRound := msg.FromDagRound
		nextLeader := p2p.ID(msg.LeaderAddress)

		if nextRound <= atomic.LoadInt64(&primR.CurrentDagRound) && len(primR.leader) == 0 {
			primR.leader = nextLeader
			primR.NextLeader = nextLeader
			primR.NextRoundToLeaderChange = nextRound
		}

	//TODO: better to use subscribe model for parent request. for follow up faster
	case *primaryproto.PrimaryParentReq:

		parents := make([][]byte, 0)
		targetRound := msg.DagRound
		// currently only support last round and current round
		// if request is current round, send current parents which is not convinced completely
		if targetRound == primR.CurrentDagRound || targetRound == 0 {
			primR.roundMtx.RLock()
			for _, v := range primR.CurrentParents {
				parents = append(parents, v.ToBytes())
			}
			targetRound = primR.CurrentDagRound
			primR.roundMtx.RUnlock()
		} else if targetRound == primR.CurrentDagRound-1 {
			primR.roundMtx.RLock()
			for _, v := range primR.LastParents {
				parents = append(parents, v.ToBytes())
			}
			primR.roundMtx.RUnlock()
		} else {
			primR.Logger.Error("invalid parent request", "req", msg)
			return
		}

		resp := primaryproto.PrimaryParentResp{
			DagRound:   targetRound,
			CertDigest: parents,
		}

		success := e.Src.Send(p2p.Envelope{
			ChannelID: primaryChannel,
			Message:   &resp,
		})
		if !success {
			//TODO: retry
			primR.Logger.Error("failed to send vode", "src", e.Src)
		}

	case *primaryproto.PrimaryParentResp:

		if len(msg.CertDigest) == 0 {
			primR.Logger.Error("received invalid parent response", "resp", msg)
			return
		}

		//TODO: lock
		if msg.DagRound == primR.CurrentDagRound {
			for i, d := range msg.CertDigest {
				var digest types.DagCertKey
				copy(digest[:], d)
				id := p2p.ID(msg.OriginAddress[i])
				primR.CurrentParents[id] = digest
			}
		} else if msg.DagRound == primR.CurrentDagRound-1 {
			for i, d := range msg.CertDigest {
				var digest types.DagCertKey
				copy(digest[:], d)
				id := p2p.ID(msg.OriginAddress[i])
				primR.LastParents[id] = digest
			}
		}

	default:
		primR.Logger.Error("unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		primR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
		return
	}

	// broadcasting happens from go routines per peer
}

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// TxsMessage is a Message containing transactions.
type TxsMessage struct {
	Txs []types.Tx
}

// String returns a string representation of the TxsMessage.
func (m *TxsMessage) String() string {
	return fmt.Sprintf("[TxsMessage %v]", m.Txs)
}
