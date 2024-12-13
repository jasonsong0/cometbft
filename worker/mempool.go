package worker

import (
	"crypto/sha256"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/types"
)

// mempool in worker is get Txs from routing mempool
// no interacts with peers but interacts with application to checkTx

//go:generate ../scripts/mockery_generate.sh Mempool

// Mempool defines the mempool interface.
//
// Updates to the mempool need to be synchronized with committing a block so
// applications can reset their transient state on Commit.
type Mempool interface {
	// CheckTx executes a new transaction against the application to determine
	// its validity and whether it should be added to the mempool.
	CheckTx(tx types.Tx, callback func(*abci.ResponseCheckTx), txInfo TxInfo) error

	// RemoveTxByKey removes a transaction, identified by its key,
	// from the mempool.
	RemoveTxByKey(txKey types.TxKey) error

	// ReapMaxBytesMaxGas reaps transactions from the mempool up to maxBytes
	// bytes total with the condition that the total gasWanted must be less than
	// maxGas.
	//
	// If both maxes are negative, there is no cap on the size of all returned
	// transactions (~ all available transactions).
	ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs

	// ReapMaxTxs reaps up to max transactions from the mempool. If max is
	// negative, there is no cap on the size of all returned transactions
	// (~ all available transactions).
	ReapMaxTxs(max int) types.Txs

	// Lock locks the mempool. The consensus must be able to hold lock to safely
	// update.
	// Before acquiring the lock, it signals the mempool that a new update is coming.
	// If the mempool is still rechecking at this point, it should be considered full.
	Lock()

	// Unlock unlocks the mempool.
	Unlock()

	// Update informs the mempool that the given txs were made for a batch and can be discarded.
	// The information will be passed to the batch
	//
	// Lock/Unlock must be managed by the caller.
	Update(
		blockHeight int64,
		blockTxs types.Txs,
		deliverTxResponses []*abci.ExecTxResult,
		newPreFn PreCheckFunc,
		newPostFn PostCheckFunc,
	) error

	// Flush removes all transactions from the mempool and caches.
	Flush()

	// Size returns the number of transactions in the mempool.
	Size() int

	// SizeBytes returns the total size of all txs in the mempool.
	SizeBytes() int64
}

// PreCheckFunc is an optional filter executed before CheckTx and rejects
// transaction if false is returned. An example would be to ensure that a
// transaction doesn't exceeded the block size.
type PreCheckFunc func(types.Tx) error

// PostCheckFunc is an optional filter executed after CheckTx and rejects
// transaction if false is returned. An example would be to ensure a
// transaction doesn't require more gas than available for the block.
type PostCheckFunc func(types.Tx, *abci.ResponseCheckTx) error

// PreCheckMaxBytes checks that the size of the transaction is smaller or equal
// to the expected maxBytes.
func PreCheckMaxBytes(maxBytes int64) PreCheckFunc {
	return func(tx types.Tx) error {
		txSize := types.ComputeProtoSizeForTxs([]types.Tx{tx})

		if txSize > maxBytes {
			return fmt.Errorf("tx size is too big: %d, max: %d", txSize, maxBytes)
		}

		return nil
	}
}

// PostCheckMaxGas checks that the wanted gas is smaller or equal to the passed
// maxGas. Returns nil if maxGas is -1.
func PostCheckMaxGas(maxGas int64) PostCheckFunc {
	return func(tx types.Tx, res *abci.ResponseCheckTx) error {
		if maxGas == -1 {
			return nil
		}
		if res.GasWanted < 0 {
			return fmt.Errorf("gas wanted %d is negative",
				res.GasWanted)
		}
		if res.GasWanted > maxGas {
			return fmt.Errorf("gas wanted %d is greater than max gas %d",
				res.GasWanted, maxGas)
		}

		return nil
	}
}

// TxKey is the fixed length array key used as an index.
type TxKey [sha256.Size]byte
