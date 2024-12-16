package worker

import (
	"fmt"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
)

// rename IDs to IDXs. not to confuse with the peer.ID
type workerIDXs struct {
	mtx        cmtsync.RWMutex
	peerMap    map[p2p.ID]uint16
	nextIDX    uint16              // assumes that a node will never have over 65536 active peers
	activeIDXs map[uint16]struct{} // used to check if a given peerID key is used, the value doesn't matter
}

// Reserve searches for the next unused ID and assigns it to the
// peer.
func (ids *workerIDXs) ReserveForPeer(peer p2p.Peer) {
	ids.mtx.Lock()
	defer ids.mtx.Unlock()

	curIDX := ids.nextPeerIDX()
	ids.peerMap[peer.ID()] = curIDX
	ids.activeIDXs[curIDX] = struct{}{}
}

// nextPeerID returns the next unused peer ID to use.
// This assumes that ids's mutex is already locked.
func (ids *workerIDXs) nextPeerIDX() uint16 {
	if len(ids.activeIDXs) == MaxActiveIDs {
		panic(fmt.Sprintf("node has maximum %d active IDs and wanted to get one more", MaxActiveIDs))
	}

	_, idExists := ids.activeIDXs[ids.nextIDX]
	for idExists {
		ids.nextIDX++
		_, idExists = ids.activeIDXs[ids.nextIDX]
	}
	curIDX := ids.nextIDX
	ids.nextIDX++
	return curIDX
}

// Reclaim returns the ID reserved for the peer back to unused pool.
func (ids *workerIDXs) Reclaim(peer p2p.Peer) {
	ids.mtx.Lock()
	defer ids.mtx.Unlock()

	removedID, ok := ids.peerMap[peer.ID()]
	if ok {
		delete(ids.activeIDXs, removedID)
		delete(ids.peerMap, peer.ID())
	}
}

// GetForPeer returns an ID reserved for the peer.
func (ids *workerIDXs) GetIdxForPeer(peer p2p.Peer) uint16 {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	return ids.peerMap[peer.ID()]
}

// GetForPeer returns an ID reserved for the peer.
func (ids *workerIDXs) GetIdxForPeerByID(peerID p2p.ID) uint16 {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	return ids.peerMap[peerID]
}

func newWorkerIDXs() *workerIDXs {
	return &workerIDXs{
		peerMap:    make(map[p2p.ID]uint16),
		activeIDXs: map[uint16]struct{}{0: {}},
		nextIDX:    1, // reserve unknownPeerID(0) for mempoolReactor.BroadcastTx
	}
}
