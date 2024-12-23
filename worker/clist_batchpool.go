package worker

import (
	"sync"
	"sync/atomic"

	"github.com/cometbft/cometbft/libs/clist"
	"github.com/cometbft/cometbft/libs/log"
	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/types"
)

type ReapType int

const (
	ReapNormal ReapType = iota
	ReapMaxSize
	ReapMaxGas
)

// CListBatchpool is an ordered in-memory pool for batchs before they are
// proposed in a consensus round. The Batchpool uses a concurrent list structure for storing batchs that can
// be efficiently accessed by multiple concurrent readers.
type CListBatchpool struct {
	height       atomic.Int64 // the last block Update()'d to
	batchesBytes atomic.Int64 // total size of Batchpool, in bytes

	// notify listeners (ie. primary) when batchs are available
	notifiedBatchAvailable atomic.Bool
	batchesAvailable       chan struct{} // fires once for each dag-round, when the Batchpool is not empty

	//TODO: make a config for batch
	//config *config.BatchpoolConfig

	//TODO: In poc v1, we use only 1 worker. So we don't need to consider gas in batchpool.
	// If multiple workers are making batchs, we need to consider gas when make batchs.
	// UpdateMtx is only used to update batchs
	updateMtx cmtsync.RWMutex

	batches *clist.CList // concurrent linked-list of batchs

	//TODO(jason): need to add AppConn for gc and other functions. app conn for mempool can be used.
	//             flush abci messages can be used for gc.
	//proxyAppConn proxy.AppConnBatchpool

	// Map for quick access to txs to record sender in CheckTx.
	// batchKey: batchKey -> CElement
	batchesMap sync.Map

	logger  log.Logger
	metrics *Metrics
}

var _ Batchpool = &CListBatchpool{}

// CListBatchpoolOption sets an optional parameter on the Batchpool.
type CListBatchpoolOption func(*CListBatchpool)

// NewCListBatchpool returns a new Batchpool with the given configuration and
// connection to an application.
func NewCListBatchpool(
	//cfg *config.BatchpoolConfig,		//TODO: add config for batchpool later
	//proxyAppConn proxy.AppConnBatchpool,	//TODO: need to connect app later
	height int64,
	options ...CListBatchpoolOption,
) *CListBatchpool {
	mp := &CListBatchpool{
		//config:       cfg,
		//proxyAppConn: proxyAppConn,
		batches: clist.New(),
		logger:  log.NewNopLogger(),
		metrics: NopMetrics(),
	}
	mp.height.Store(height)

	//TODO: need to connect app later
	//proxyAppConn.SetResponseCallback(mp.globalCb)

	//NOTE: only support metric option now
	for _, option := range options {
		option(mp)
	}

	return mp
}

func (mem *CListBatchpool) getCElement(bKey types.BatchKey) (*clist.CElement, bool) {
	if e, ok := mem.batchesMap.Load(bKey); ok {
		return e.(*clist.CElement), true
	}
	return nil, false
}

func (mem *CListBatchpool) getMemTx(bKey types.BatchKey) *batchpoolBatch {
	if e, ok := mem.getCElement(bKey); ok {
		return e.Value.(*batchpoolBatch)
	}
	return nil
}

func (mem *CListBatchpool) removeAllBatches() {
	for e := mem.batches.Front(); e != nil; e = e.Next() {
		mem.batches.Remove(e)
		e.DetachPrev()
	}

	mem.batchesMap.Range(func(key, _ interface{}) bool {
		mem.batchesMap.Delete(key)
		return true
	})
}

// NOTE: not thread safe - should only be called once, on startup
func (mem *CListBatchpool) EnableBatchesAvailable() {
	mem.batchesAvailable = make(chan struct{}, 1)
}

// SetLogger sets the Logger.
func (mem *CListBatchpool) SetLogger(l log.Logger) {
	mem.logger = l
}

// WithMetrics sets the metrics.
func WithBatchMetrics(metrics *Metrics) CListBatchpoolOption {
	return func(mem *CListBatchpool) { mem.metrics = metrics }
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) Lock() {
	mem.updateMtx.Lock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) Unlock() {
	mem.updateMtx.Unlock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) Size() int {
	return mem.batches.Len()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) SizeBytes() int64 {
	return mem.batchesBytes.Load()
}

// Lock() must be help by the caller during execution.
func (mem *CListBatchpool) FlushAppConn() error {
	/*
		//TODO: add for gc
		err := mem.proxyAppConn.Flush(context.TODO())
		if err != nil {
			return ErrFlushAppConn{Err: err}
		}
	*/

	return nil
}

// XXX: Unsafe! Calling Flush may leave Batchpool in inconsistent state.
func (mem *CListBatchpool) Flush() {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	mem.batchesBytes.Store(0)

	mem.removeAllBatches()
}

// TxsFront returns the first transaction in the ordered list for peer
// goroutines to call .NextWait() on.
// FIXME: leaking implementation details!
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) BatchesFront() *clist.CElement {
	return mem.batches.Front()
}

// TxsWaitChan returns a channel to wait on transactions. It will be closed
// once the Batchpool is not empty (ie. the internal `mem.txs` has at least one
// element)
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) BatchesWaitChan() <-chan struct{} {
	return mem.batches.WaitChan()
}

//TODO: check if globalCB from abci is needed

// Called from:
//   - resCbFirstTime (lock not held) if tx is valid
func (mem *CListBatchpool) addBatch(b *batchpoolBatch) {
	e := mem.batches.PushBack(b)
	mem.batchesMap.Store(b.batch.BatchKey, e)
	mem.batchesBytes.Add(b.batch.BatchSize)
	mem.metrics.TxSizeBytes.Observe(float64(b.batch.BatchSize))

	//TODO: store to dagstore

}

// RemoveTxByKey removes a transaction from the Batchpool by its TxKey index.
// Called from:
//   - Update (lock held) if tx was committed
//   - resCbRecheck (lock not held) if tx was invalidated
func (mem *CListBatchpool) RemoveBatchByKey(bKey types.BatchKey) error {
	if elem, ok := mem.getCElement(bKey); ok {
		mem.batches.Remove(elem)
		elem.DetachPrev()
		mem.batchesMap.Delete(bKey)
		batch := elem.Value.(*batchpoolBatch).batch
		mem.batchesBytes.Add(int64(-batch.BatchSize))
		return nil
	}
	return ErrTxNotFound
}

func (mem *CListBatchpool) isFull(txSize int) error {

	//TODO: check batchpool is full or not
	return nil
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) BatchesAvailable() <-chan struct{} {
	return mem.batchesAvailable
}

func (mem *CListBatchpool) notifyBatchesAvailable() {
	if mem.Size() == 0 {
		panic("notified batches available but Batchpool is empty!")
	}
	if mem.batchesAvailable != nil && mem.notifiedBatchAvailable.CompareAndSwap(false, true) {
		// channel cap is 1, so this will send once
		select {
		case mem.batchesAvailable <- struct{}{}:
		default:
		}
	}
}

// Safe for concurrent use by multiple goroutines.
// primary decides mas bytes, gas for each worker.
// only returns acked batches
// if return over size, trigger type = 1
// if return over gas, trigger type = 2
// else trigger type = 0
func (mem *CListBatchpool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) (types.Batches, ReapType) {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	var (
		totalGas    int64
		runningSize int64
	)

	// TODO: we will get a performance boost if we have a good estimate of avg
	// size per tx, and set the initial capacity based off of that.
	// txs := make([]types.Tx, 0, cmtmath.MinInt(mem.txs.Len(), max/mem.avgTxSize))
	batches := make([]types.Batch, 0, mem.batches.Len())
	for e := mem.batches.Front(); e != nil; e = e.Next() {
		b := e.Value.(*batchpoolBatch)

		// should be acked to process next step
		if !b.isMaj23() {
			break
		}

		batches = append(batches, b.batch)

		dataSize := types.ComputeProtoSizeForBatches([]types.Batch{b.batch})

		// Check total size requirement
		if maxBytes > -1 && runningSize+int64(dataSize) > maxBytes {
			return batches[:len(batches)-1], ReapMaxSize
		}

		runningSize += int64(dataSize)

		// Check total gas requirement.
		// If maxGas is negative, skip this check.
		// Since newTotalGas < masGas, which
		// must be non-negative, it follows that this won't overflow.
		newTotalGas := totalGas + b.gasWanted
		if maxGas > -1 && newTotalGas > maxGas {
			return batches[:len(batches)-1], ReapMaxGas
		}
		totalGas = newTotalGas
	}
	return batches, ReapNormal
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListBatchpool) ReapMaxBatches(max int) types.Batches {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	if max < 0 {
		max = mem.batches.Len()
	}

	batches := make([]types.Batch, 0, cmtmath.MinInt(mem.batches.Len(), max))
	for e := mem.batches.Front(); e != nil && len(batches) <= max; e = e.Next() {
		b := e.Value.(*batchpoolBatch)
		batches = append(batches, b.batch)
	}
	return batches
}

// Lock() must be help by the caller during execution.
func (mem *CListBatchpool) Update(
	height int64,
	batches types.Batches,
) error {
	mem.logger.Debug("Update", "height", height, "len(batches)", len(batches))

	// Set height
	mem.height.Store(height)
	mem.notifiedBatchAvailable.Store(false)

	for _, batch := range batches {
		// Remove committed batch from the Batchpool.
		//
		if err := mem.RemoveBatchByKey(batch.BatchKey); err != nil {
			mem.logger.Debug("Committed batch not in local Batchpool (not an error)",
				"key", batch.BatchKey,
				"error", err.Error())
		}
	}

	// Notify if there are still batches left in the Batchpool.
	if mem.Size() > 0 {
		mem.notifyBatchesAvailable()
	}

	// Update metrics
	mem.metrics.Size.Set(float64(mem.Size()))
	mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))

	return nil
}

// called when batchpool consumpted by primary
func (bp *CListBatchpool) UpdateByPrimary(b types.Batch) error {
	bp.logger.Debug("UpdateByPrimary", "batchkey", b.BatchKey, "size", b.BatchSize)

	// Remove committed batch from the Batchpool.
	//
	if err := bp.RemoveBatchByKey(b.BatchKey); err != nil {
		bp.logger.Debug("Committed batch not in local Batchpool (not an error)",
			"key", b.BatchKey,
			"error", err.Error())
	}

	return nil
}
