package types

import (
	"crypto/sha256"
	"fmt"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// BatchHashSize is the size of the batch digset key index
const BatchKeySize = sha256.Size

type BatchKey [BatchKeySize]byte

type Batch struct {

	// Batch is an Tx array
	TxArr Txs

	// BatchKey is the fixed length array key used as an index.
	BatchKey [BatchKeySize]byte

	BatchOwner string //TODO: p2pID + workerID

	BatchSize int64

	BatchCreated time.Time
	BatchRecved  time.Time // recv time at reactor
}

type Batches []Batch

// Hash computes the TMHASH hash of the wire encoded transaction.
func (b Batch) Hash() []byte {
	return b.TxArr.Hash()
}

// String returns the hex-encoded batch as a string.
func (b Batch) String() string {
	return fmt.Sprintf("Batch{%X} from %s, %x->%x", b.BatchKey, b.BatchOwner, b.BatchCreated, b.BatchRecved)
}

// ToProto converts Data to protobuf
func (b *Batch) ToProto() cmtproto.Batch {
	tp := new(cmtproto.Batch)

	tp.BatchKey = b.BatchKey[:]
	tp.BatchOwner = b.BatchOwner
	tp.BatchSize = b.BatchSize
	tp.BatchCreated = b.BatchCreated.Unix()
	tp.BatchRecved = b.BatchRecved.Unix()

	if len(b.TxArr) > 0 {
		txBzs := make([][]byte, len(b.TxArr))
		for i := range b.TxArr {
			txBzs[i] = b.TxArr[i]
		}
		tp.Txs = txBzs
	}

	return *tp
}

func ComputeProtoSizeForBatches(batches []Batch) int64 {
	sum := int64(0)
	for _, b := range batches {
		sum = sum + b.ToProto().Size()
	}
	return sum
}
