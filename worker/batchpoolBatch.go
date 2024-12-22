package worker

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/types"
)

// batchpoolBatch is an entry in the batchpool
type batchpoolBatch struct {
	height    int64              // height that this tx had been created in
	gasWanted int64              // amount of gas this batch states it will require
	batch     types.Batch        // validated by the application
	created   time.Time          // time when this batch was created. if time is set, it means already pushed to the primary
	acked     map[uint16]float64 // peerIDX -> voting power
	maj23     bool               // whether this batch has been acked by maj23

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

// TODO: change to voting power of past dag round
func (bBatch *batchpoolBatch) isMaj23() bool {

	const minVoteNodeCnt = 2

	// Set height
	sumVotingPower := float64(0)
	for _, votingPower := range bBatch.acked {
		sumVotingPower += votingPower
	}

	//TODO: change the logic. its temporary for poc v1.
	if len(bBatch.acked) > minVoteNodeCnt && sumVotingPower >= 2.0 {
		return true
	}

	return false
}
