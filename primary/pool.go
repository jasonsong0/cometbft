package primary

import (
	"math"

	primaryproto "github.com/cometbft/cometbft/proto/tendermint/primary"
	"github.com/cometbft/cometbft/types"
)

// primary pool used for queueing and processing header
// only meant for leader node

const (
	PrimarypoolChannel = byte(0x31) // header channel used 0x31 for poc v1

	// PeerCatchupSleepIntervalMS defines how much time to sleep if a peer is behind
	PeerCatchupSleepIntervalMS = 100

	// UnknownPeerID is the peer ID to use when running CheckTx when there is
	// no peer (e.g. RPC)
	UnknownPeerID uint16 = 0

	MaxActiveIDs = math.MaxUint16
)

//go:generate ../scripts/mockery_generate.sh Mempool

// primary pool contains the interface to handling header
// all header in the pool should be certified
type pool interface {
	AddHeader(header primaryproto.PrimaryDagHeader) error
	RemoveHeaderByKey(batchKey types.DagHeaderKey) error

	Lock()
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
		blockHeaderKey types.DagHeaderKey,
	) error
	// TODO: check garbage collection logic

	// Flush removes all headers from the pool
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
