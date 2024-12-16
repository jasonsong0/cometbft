package worker

import (
	"math"

	"github.com/cometbft/cometbft/types"
)

const (
	BatchpoolChannelBase = byte(0x30) // Batch channel used 0x30~0x3f for poc v1

	// PeerCatchupSleepIntervalMS defines how much time to sleep if a peer is behind
	PeerCatchupSleepIntervalMS = 100

	// UnknownPeerID is the peer ID to use when running CheckTx when there is
	// no peer (e.g. RPC)
	UnknownPeerID uint16 = 0

	MaxActiveIDs = math.MaxUint16
)

//go:generate ../scripts/mockery_generate.sh Mempool

// batchpool defines the batchpool interface.
//
// Updates to the batchpool need to be synchronized with committing a block so
// applications can reset their transient state on Commit.
// application(or consensus) can request garbage collection for clear old batches.
type Batchpool interface {

	// RemoveTxByKey removes a batch, identified by its key,
	// from the batchpool.
	RemoveBatchByKey(batchKey types.BatchKey) error

	// ReapMaxBytesMaxGas reaps batchs from the batchpool up to maxBytes
	// bytes total with the condition that the total gasWanted must be less than
	// maxGas.
	//
	// If both maxes are negative, there is no cap on the size of all returned
	// transactions (~ all available transactions).
	ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Batches

	// Lock locks the batchpool. The consensus must be able to hold lock to safely
	// update.
	// Before acquiring the lock, it signals the batchpool that a new update is coming.
	// If the batchpool is still rechecking at this point, it should be considered full.
	Lock()

	// Unlock unlocks the batchpool.
	Unlock()

	// Update informs the batchpool that the given batches were committed and can be
	// discarded.
	//
	// NOTE:
	// 1. This should be called *after* block is committed by consensus.
	// 2. Lock/Unlock must be managed by the caller.
	// TODO: check batch
	Update(
		blockHeight int64,
		blockBatches types.Batches,
	) error
	// TODO: check garbage collection logic

	// Flush removes all transactions from the batchpool and caches.
	Flush()

	// TxsAvailable returns a channel which fires once for every dag-round, and only
	// when batches are available in the batchpool.
	//
	// NOTE:
	// 1. The returned channel may be nil if EnableBatchesAvailable was not called.
	BatchesAvailable() <-chan struct{}

	// EnableTxsAvailable initializes the TxsAvailable channel, ensuring it will
	// trigger once every height when transactions are available.
	EnableBatchesAvailable()

	// Size returns the number of batches in the batchpool.
	Size() int

	// SizeBytes returns the total size of all batches in the batchpool.
	SizeBytes() int64
}
