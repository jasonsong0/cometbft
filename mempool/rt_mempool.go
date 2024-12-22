package mempool

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/worker"
)

// RtMempool is a mempool that forward tx to worker.
// cacheing for filtering dup TXs
//
// The Worker is responsible for storing, disseminating
// The Primary is reponsible for proposing transactions(makeing header).

// CListMempool is an ordered in-memory pool for transactions before they are
// proposed in a consensus round. Transaction validity is checked using the
// CheckTx abci message before the transaction is added to the pool. The
// mempool uses a concurrent list structure for storing transactions that can
// be efficiently accessed by multiple concurrent readers.
type RtMempool struct {

	//config *config.MempoolConfig

	// Exclusive mutex for Update method to prevent concurrent execution of
	// CheckTx or ReapMaxBytesMaxGas(ReapMaxTxs) methods.
	updateMtx cmtsync.RWMutex

	//TODO: for v1. use single worker.
	//routing the tx directly
	workerMempool worker.Mempool

	logger  log.Logger
	metrics *Metrics
}

var _ Mempool = &RtMempool{}

// mempool CheckTx called by client. bypass to worker CheckTx.
func (mem *RtMempool) CheckTx(
	tx types.Tx,
	cb func(*abci.ResponseCheckTx),
	txInfo TxInfo) error {

	//TODO: routing to worker
	ti := worker.TxInfo{
		SenderID:    txInfo.SenderID,
		SenderP2PID: txInfo.SenderP2PID,
	}
	mem.workerMempool.CheckTx(tx, cb, ti)

	return nil

}

// RemoveTxByKey always returns an error.
func (*RtMempool) RemoveTxByKey(types.TxKey) error {
	panic("rtMempool RemoveTxByKey should not be called")
	//return errNotAllowed
}

// ReapMaxBytesMaxGas always returns nil.
func (*RtMempool) ReapMaxBytesMaxGas(int64, int64) types.Txs {
	panic("rtMempool ReapMAxBytesMaxGas should not be called")
	//return nil
}

// ReapMaxTxs always returns nil.
func (*RtMempool) ReapMaxTxs(int) types.Txs {
	panic("rtMempool ReapMAxTxs should not be called")
	//return nil
}

// Lock does nothing.
func (*RtMempool) Lock() {}

// Unlock does nothing.
func (*RtMempool) Unlock() {}

// Update does nothing.
func (*RtMempool) Update(
	int64,
	types.Txs,
	[]*abci.ExecTxResult,
	PreCheckFunc,
	PostCheckFunc,
) error {
	return nil
}

// FlushAppConn does nothing.
func (*RtMempool) FlushAppConn() error { return nil }

// Flush does nothing.
func (*RtMempool) Flush() {}

// TxsAvailable always returns nil.
func (*RtMempool) TxsAvailable() <-chan struct{} {
	return nil
}

// EnableTxsAvailable does nothing.
func (*RtMempool) EnableTxsAvailable() {}

// SetTxRemovedCallback does nothing.
func (*RtMempool) SetTxRemovedCallback(func(txKey types.TxKey)) {}

// Size always returns 0.
func (*RtMempool) Size() int { return 0 }

// SizeBytes always returns 0.
func (*RtMempool) SizeBytes() int64 { return 0 }

// RtMempoolReactor is a mempool reactor that does nothing.
type RtMempoolReactor struct {
	service.BaseService
}

// NewRtMempoolReactor returns a new `rt` reactor.
//
// To be used only in RPC.
func NewRtMempoolReactor() *RtMempoolReactor {
	return &RtMempoolReactor{*service.NewBaseService(nil, "RtMempoolReactor", nil)}
}

var _ p2p.Reactor = &RtMempoolReactor{}

// GetChannels always returns nil.
func (*RtMempoolReactor) GetChannels() []*p2p.ChannelDescriptor { return nil }

// AddPeer does nothing.
func (*RtMempoolReactor) AddPeer(p2p.Peer) {}

// InitPeer always returns nil.
func (*RtMempoolReactor) InitPeer(p2p.Peer) p2p.Peer { return nil }

// RemovePeer does nothing.
func (*RtMempoolReactor) RemovePeer(p2p.Peer, interface{}) {}

// Receive does nothing.
func (*RtMempoolReactor) Receive(p2p.Envelope) {}

// SetSwitch does nothing.
func (*RtMempoolReactor) SetSwitch(*p2p.Switch) {}
