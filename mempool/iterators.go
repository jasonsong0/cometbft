package mempool

import (
	"fmt"

	"github.com/cometbft/cometbft/internal/clist"
	"github.com/cometbft/cometbft/types"
)

// IWRRIterator is the base struct for implementing iterators that traverse lanes with
// the Interleaved Weighted Round Robin (WRR) algorithm.
// https://en.wikipedia.org/wiki/Weighted_round_robin
type IWRRIterator struct {
	sortedLanes []types.Lane
	laneIndex   int                            // current lane being iterated; index on sortedLanes
	cursors     map[types.Lane]*clist.CElement // last accessed entries on each lane
	round       int                            // counts the rounds for IWRR
}

// This function picks the next lane to fetch an item from.
// If it was the last lane, it advances the round counter as well.
func (iter *IWRRIterator) advanceIndexes() types.Lane {
	if iter.laneIndex == len(iter.sortedLanes)-1 {
		iter.round = (iter.round + 1) % (int(iter.sortedLanes[0]) + 1)
		if iter.round == 0 {
			iter.round++
		}
	}
	iter.laneIndex = (iter.laneIndex + 1) % len(iter.sortedLanes)
	return iter.sortedLanes[iter.laneIndex]
}

// Non-blocking version of the IWRR iterator to be used for reaping and
// rechecking transactions.
//
// This iterator does not support changes on the underlying mempool once initialized (or `Reset`),
// therefore the lock must be held on the mempool when iterating.
type NonBlockingIterator struct {
	IWRRIterator
}

func NewNonBlockingIterator(mem *CListMempool) *NonBlockingIterator {
	baseIter := IWRRIterator{
		sortedLanes: mem.sortedLanes,
		cursors:     make(map[types.Lane]*clist.CElement, len(mem.lanes)),
		round:       1,
	}
	iter := &NonBlockingIterator{
		IWRRIterator: baseIter,
	}
	iter.reset(mem.lanes)
	return iter
}

// Reset must be called before every use of the iterator.
func (iter *NonBlockingIterator) reset(lanes map[types.Lane]*clist.CList) {
	iter.laneIndex = 0
	iter.round = 1
	// Set cursors at the beginning of each lane.
	for lane := range lanes {
		iter.cursors[lane] = lanes[lane].Front()
	}
}

// Next returns the next element according to the WRR algorithm.
func (iter *NonBlockingIterator) Next() Entry {
	numEmptyLanes := 0

	lane := iter.sortedLanes[iter.laneIndex]
	for {
		// Skip empty lane or if cursor is at end of lane.
		if iter.cursors[lane] == nil {
			numEmptyLanes++
			if numEmptyLanes >= len(iter.sortedLanes) {
				return nil
			}
			lane = iter.advanceIndexes()
			continue
		}
		// Skip over-consumed lane on current round.
		if int(lane) < iter.round {
			numEmptyLanes = 0
			lane = iter.advanceIndexes()
			continue
		}
		break
	}
	elem := iter.cursors[lane]
	if elem == nil {
		panic(fmt.Errorf("Iterator picked a nil entry on lane %d", lane))
	}
	iter.cursors[lane] = iter.cursors[lane].Next()
	_ = iter.advanceIndexes()
	return elem.Value.(*mempoolTx)
}

// BlockingIterator implements a blocking version of the WRR iterator,
// meaning that when no transaction is available, it will wait until a new one
// is added to the mempool.
// Unlike `NonBlockingIterator`, this iterator is expected to work with an evolving mempool.
type BlockingIterator struct {
	IWRRIterator
	mp *CListMempool
}

func NewBlockingIterator(mem *CListMempool) Iterator {
	iter := IWRRIterator{
		sortedLanes: mem.sortedLanes,
		cursors:     make(map[types.Lane]*clist.CElement, len(mem.sortedLanes)),
		round:       1,
	}
	return &BlockingIterator{
		IWRRIterator: iter,
		mp:           mem,
	}
}

// WaitNextCh returns a channel to wait for the next available entry. The channel will be explicitly
// closed when the entry gets removed before it is added to the channel, or when reaching the end of
// the list.
//
// Unsafe for concurrent use by multiple goroutines.
func (iter *BlockingIterator) WaitNextCh() <-chan Entry {
	ch := make(chan Entry)
	go func() {
		var lane types.Lane
		for {
			l, addTxCh := iter.pickLane()
			if addTxCh == nil {
				lane = l
				break
			}
			// There are no transactions to take from any lane. Wait until at
			// least one is added to the mempool and try again.
			<-addTxCh
		}
		if elem := iter.next(lane); elem != nil {
			ch <- elem.Value.(Entry)
		}
		// Unblock receiver in case no entry was sent (it will receive nil).
		close(ch)
	}()
	return ch
}

// pickLane returns a _valid_ lane on which to iterate, according to the WRR
// algorithm. A lane is valid if it is not empty and it is not over-consumed,
// meaning that the number of accessed entries in the lane has not yet reached
// its priority value in the current WRR iteration. It returns a channel to wait
// for new transactions if all lanes are empty or don't have transactions that
// have not yet been accessed.
func (iter *BlockingIterator) pickLane() (types.Lane, chan struct{}) {
	iter.mp.addTxChMtx.RLock()
	defer iter.mp.addTxChMtx.RUnlock()

	// Start from the last accessed lane.
	lane := iter.sortedLanes[iter.laneIndex]

	// Loop until finding a valid lane. If the current lane is not valid,
	// continue with the next lower-priority lane, in a round robin fashion.
	numEmptyLanes := 0
	for {
		// Skip empty lanes or lanes with their cursor pointing at their last entry.
		if iter.mp.lanes[lane].Len() == 0 ||
			(iter.cursors[lane] != nil &&
				iter.cursors[lane].Value.(*mempoolTx).seq == iter.mp.addTxLaneSeqs[lane]) {
			numEmptyLanes++
			if numEmptyLanes >= len(iter.sortedLanes) {
				// There are no lanes with non-accessed entries. Wait until a
				// new tx is added.
				return 0, iter.mp.addTxCh
			}
			lane = iter.advanceIndexes()
			continue
		}

		// Skip over-consumed lanes.
		if int(lane) < iter.round {
			numEmptyLanes = 0
			lane = iter.advanceIndexes()
			continue
		}

		_ = iter.advanceIndexes()
		return lane, nil
	}
}

// next returns the next entry from the given lane and updates WRR variables.
func (iter *BlockingIterator) next(lane types.Lane) *clist.CElement {
	// Load the last accessed entry in the lane and set the next one.
	var next *clist.CElement
	if cursor := iter.cursors[lane]; cursor != nil {
		// If the current entry is the last one or was removed, Next will return nil.
		// Note we don't need to wait until the next entry is available (with <-cursor.NextWaitChan()).
		next = cursor.Next()
	} else {
		// We are at the beginning of the iteration or the saved entry got removed. Pick the first
		// entry in the lane if it's available (don't wait for it); if not, Front will return nil.
		next = iter.mp.lanes[lane].Front()
	}

	// Update auxiliary variables.
	if next != nil {
		// Save entry.
		iter.cursors[lane] = next
	} else {
		// The entry got removed or it was the last one in the lane.
		// At the moment this should not happen - the loop in PickLane will loop forever until there
		// is data in at least one lane
		delete(iter.cursors, lane)
	}

	return next
}