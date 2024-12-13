package worker

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/types"
)

// batchpoolBatch is an entry in the batchpool
type batchpoolBatch struct {
	height    int64       // height that this tx had been created in
	gasWanted int64       // amount of gas this batch states it will require
	batch     types.Batch // validated by the application
	created   time.Time   // time when this batch was created

	senders sync.Map
}

// Height returns the height for this transaction
func (bBatch *batchpoolBatch) Height() int64 {
	return atomic.LoadInt64(&bBatch.height)
}

// Time returns the time when this batch was created
func (bBatch *batchpoolBatch) Time() time.Time {
	return bBatch.created
}

func (bBatch *batchpoolBatch) isSender(peerID uint16) bool {
	_, ok := bBatch.senders.Load(peerID)
	return ok
}

func (bBatch *batchpoolBatch) addSender(senderID uint16) bool {
	_, added := bBatch.senders.LoadOrStore(senderID, true)
	return added
}
