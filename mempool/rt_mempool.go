package mempool

import (
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
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
	height   atomic.Int64 // the last block Update()'d to
	
	// notify listeners (ie. consensus) when txs are available
	notifiedTxsAvailable atomic.Bool
	txsAvailable         chan struct{} // fires once for each height, when the mempool is not empty

	//config *config.MempoolConfig

	// Exclusive mutex for Update method to prevent concurrent execution of
	// CheckTx or ReapMaxBytesMaxGas(ReapMaxTxs) methods.
	updateMtx cmtsync.RWMutex
	preCheck  PreCheckFunc
	postCheck PostCheckFunc

	txs          *clist.CList // concurrent linked-list of good txs
	proxyAppConn proxy.AppConnMempool

	// Keeps track of the rechecking process.
	recheck *recheck

	// Map for quick access to txs to record sender in CheckTx.
	// txsMap: txKey -> CElement
	txsMap sync.Map

	// Keep a cache of already-seen txs.
	// This reduces the pressure on the proxyApp.
	cache TxCache

	logger  log.Logger
	metrics *Metrics
}

// errNotAllowed indicates that the operation is not allowed with `rt` mempool.
var errNotAllowed = errors.New("not allowed with `rt` mempool")

var _ Mempool = &RtMempool{}

// CheckTx call CheckTx to worker. 
func (mem *RtMempool) CheckTx(types.Tx, func(*abci.ResponseCheckTx), TxInfo) error {
	mem.updateMtx.RLock()
	// use defer to unlock mutex because application (*local client*) might panic
	defer mem.updateMtx.RUnlock()

	if mem.preCheck != nil {
		if err := mem.preCheck(tx); err != nil {
			return ErrPreCheck{Err: err}
		}
	}

	

	// NOTE: proxyAppConn may error if tx buffer is full
	if err := mem.proxyAppConn.Error(); err != nil {
		return ErrAppConnMempool{Err: err}
	}

	if !mem.cache.Push(tx) { // if the transaction already exists in the cache
		// Record a new sender for a tx we've already seen.
		// Note it's possible a tx is still in the cache but no longer in the mempool
		// (eg. after committing a block, txs are removed from mempool but not cache),
		// so we only record the sender for txs still in the mempool.
		if memTx := mem.getMemTx(tx.Key()); memTx != nil {
			memTx.addSender(txInfo.SenderID)
			// TODO: consider punishing peer for dups,
			// its non-trivial since invalid txs can become valid,
			// but they can spam the same tx with little cost to them atm.
		}
		return ErrTxInCache
	}

	reqRes, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.RequestCheckTx{Tx: tx})
	if err != nil {
		panic(fmt.Errorf("CheckTx request for tx %s failed: %w", log.NewLazySprintf("%v", tx.Hash()), err))
	}
	reqRes.SetCallback(mem.reqResCb(tx, txInfo, cb))

	return nil

}

// RemoveTxByKey always returns an error.
func (*RtMempool) RemoveTxByKey(types.TxKey) error { return errNotAllowed }

// ReapMaxBytesMaxGas always returns nil.
func (*RtMempool) ReapMaxBytesMaxGas(int64, int64) types.Txs { return nil }

// ReapMaxTxs always returns nil.
func (*RtMempool) ReapMaxTxs(int) types.Txs { return nil }

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
