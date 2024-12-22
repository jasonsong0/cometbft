package types

import (
	"crypto/sha256"
	"fmt"
	"time"

	workerproto "github.com/cometbft/cometbft/proto/tendermint/worker"
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

	BatchCreated   time.Time
	BatchConfirmed time.Time                // recv enough batch.
	BatchRTT       map[string]time.Duration // RTT for each peer
	BatchAcked     map[string]float64       // acked and voting power

}

type Batches []Batch

// Hash computes the TMHASH hash of the wire encoded transaction.
func (b Batch) Hash() []byte {
	return b.TxArr.Hash()
}

// String returns the hex-encoded batch as a string.
func (b Batch) String() string {
	return fmt.Sprintf("Batch{%X} from %s, %x->%x", b.BatchKey, b.BatchOwner, b.BatchCreated, b.BatchConfirmed)
}

// ToProto converts Data to protobuf
func (b *Batch) ToProto() workerproto.BatchData {
	tp := new(workerproto.BatchData)

	tp.BatchKey = b.BatchKey[:]
	tp.BatchOwner = b.BatchOwner
	tp.BatchCreated = b.BatchCreated

	if len(b.TxArr) > 0 {
		txBzs := make([][]byte, len(b.TxArr))
		for i := range b.TxArr {
			txBzs[i] = b.TxArr[i]
		}
		tp.Txs = txBzs
	}

	return *tp
}

func ComputeProtoSizeForBatches(batches []Batch) int {
	sum := int(0)
	for _, b := range batches {
		p := b.ToProto()
		sum = sum + p.Size()
	}
	return sum
}
