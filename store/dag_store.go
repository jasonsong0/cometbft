package store

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/gogoproto/proto"
	lru "github.com/hashicorp/golang-lru/v2"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cometbft/cometbft/evidence"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	cmtstore "github.com/cometbft/cometbft/proto/tendermint/store"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"
)

/*
DagStore is a simple low level store for dag Infos.

There are three types of information stored:
  - RoundVali: Validators for specific round. (not store all round. only store round that have been changed)
  - Cert:      The certification data of each block

Currently the precommit signatures are duplicated in the Block parts as
well as the Commit.  In the future this may change, perhaps by moving
the Commit data outside the Block. (TODO)

The store can be assumed to contain all contiguous blocks between base and height (inclusive).

// NOTE: DagStore methods will panic if they encounter errors
// deserializing loaded data, indicating probable corruption on disk.
*/
type DagStore struct {
	db dbm.DB

	// mtx guards access to the struct fields listed below it. Although we rely on the database
	// to enforce fine-grained concurrency control for its data, we need to make sure that
	// no external observer can get data from the database that is not in sync with the fields below,
	// and vice-versa. Hence, when updating the fields below, we use the mutex to make sure
	// that the database is also up to date. This prevents any concurrent external access from
	// obtaining inconsistent data.
	// The only reason for keeping these fields in the struct is that the data
	// can't efficiently be queried from the database since the key encoding we use is not
	// lexicographically ordered (see https://github.com/tendermint/tendermint/issues/4567).
	mtx             cmtsync.RWMutex
	base            int64 // base round. start round of store
	valiUpdateRound int64 // last round that validator set was updated
	round           int64 // current round of dag

	roundCache *lru.Cache[int64, *types.DagRound]
	//valiSetCache *lru.Cache[int64, *types.ValidatorSet]
	//certCache                *lru.Cache[int64, *types.Commit]
}

// NewDagStore returns a new DagStore with the given DB,
// initialized to the last height that was committed to the DB.
func NewDagStore(db dbm.DB) *DagStore {
	bs := LoadDagStoreState(db)
	dStore := &DagStore{
		base:  bs.Base,
		round: bs.Round,
		db:    db,
	}
	dStore.addCaches()
	return dStore
}

func (bs *DagStore) addCaches() {
	var err error
	// err can only occur if the argument is non-positive, so is impossible in context.
	bs.roundCache, err = lru.New[int64, *types.DagRound](100)
	if err != nil {
		panic(err)
	}
}

func (bs *DagStore) IsEmpty() bool {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.base == bs.round && bs.base == 0
}

// Base returns the first known contiguous block height, or 0 for empty block stores.
func (bs *DagStore) Base() int64 {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.base
}

// Height returns the last known contiguous block height, or 0 for empty block stores.
func (bs *DagStore) Round() int64 {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.round
}

// Size returns the number of blocks in the block store.
func (bs *DagStore) Size() int64 {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	if bs.round == 0 {
		return 0
	}
	return bs.round - bs.base + 1
}

// LoadBase atomically loads the base block meta, or returns nil if no base is found.
func (bs *DagStore) LoadBaseMeta() *types.DagRound {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	if bs.base == 0 {
		return nil
	}
	return bs.LoadDagRound(bs.base)
}

// LoadBlockByHash returns the block with the given hash.
// If no block is found for that hash, it returns nil.
// Panics if it fails to parse height associated with the given hash.
func (bs *DagStore) LoadBlockByHash(hash []byte) *types.Block {
	bz, err := bs.db.Get(calcBlockHashKey(hash))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}

	s := string(bz)
	height, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to extract height from %s: %v", s, err))
	}
	return bs.LoadBlock(height)
}

// LoadBlockPart returns the Part at the given index
// from the block at the given height.
// If no part is found for the given height and index, it returns nil.
func (bs *DagStore) LoadBlockPart(height int64, index int) *types.Part {
	pbpart := new(cmtproto.Part)

	bz, err := bs.db.Get(calcBlockPartKey(height, index))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}

	err = proto.Unmarshal(bz, pbpart)
	if err != nil {
		panic(fmt.Errorf("unmarshal to cmtproto.Part failed: %w", err))
	}
	part, err := types.PartFromProto(pbpart)
	if err != nil {
		panic(fmt.Sprintf("Error reading block part: %v", err))
	}

	return part
}

// LoadBlockMeta returns the BlockMeta for the given height.
// If no block is found for the given height, it returns nil.
func (bs *DagStore) LoadDagRound(round int64) *types.DagRound {
	pbbm := new(cmtproto.DagRound)
	bz, err := bs.db.Get(calcDagRoundKey(round))
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil
	}

	err = proto.Unmarshal(bz, pbbm)
	if err != nil {
		panic(fmt.Errorf("unmarshal to cmtproto.DagRound: %w", err))
	}

	d, err := types.DagRoundFromTrustedProto(pbbm)
	if err != nil {
		panic(fmt.Errorf("error from proto DagRound: %w", err))
	}

	return d
}

// LoadBlockMetaByHash returns the blockmeta who's header corresponds to the given
// hash. If none is found, returns nil.
func (bs *DagStore) LoadBlockMetaByHash(hash []byte) *types.BlockMeta {
	bz, err := bs.db.Get(calcBlockHashKey(hash))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}

	s := string(bz)
	height, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to extract height from %s: %v", s, err))
	}
	return bs.LoadBlockMeta(height)
}

// LoadBlockCommit returns the Commit for the given height.
// This commit consists of the +2/3 and other Precommit-votes for block at `height`,
// and it comes from the block.LastCommit for `height+1`.
// If no commit is found for the given height, it returns nil.
func (bs *DagStore) LoadBlockCommit(height int64) *types.Commit {
	comm, ok := bs.blockCommitCache.Get(height)
	if ok {
		return comm.Clone()
	}
	pbc := new(cmtproto.Commit)
	bz, err := bs.db.Get(calcBlockCommitKey(height))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = proto.Unmarshal(bz, pbc)
	if err != nil {
		panic(fmt.Errorf("error reading block commit: %w", err))
	}
	commit, err := types.CommitFromProto(pbc)
	if err != nil {
		panic(fmt.Errorf("converting commit to proto: %w", err))
	}
	bs.blockCommitCache.Add(height, commit)
	return commit.Clone()
}

// LoadExtendedCommit returns the ExtendedCommit for the given height.
// The extended commit is not guaranteed to contain the same +2/3 precommits data
// as the commit in the block.
func (bs *DagStore) LoadBlockExtendedCommit(height int64) *types.ExtendedCommit {
	comm, ok := bs.blockExtendedCommitCache.Get(height)
	if ok {
		return comm.Clone()
	}
	pbec := new(cmtproto.ExtendedCommit)
	bz, err := bs.db.Get(calcExtCommitKey(height))
	if err != nil {
		panic(fmt.Errorf("fetching extended commit: %w", err))
	}
	if len(bz) == 0 {
		return nil
	}
	err = proto.Unmarshal(bz, pbec)
	if err != nil {
		panic(fmt.Errorf("decoding extended commit: %w", err))
	}
	extCommit, err := types.ExtendedCommitFromProto(pbec)
	if err != nil {
		panic(fmt.Errorf("converting extended commit: %w", err))
	}
	bs.blockExtendedCommitCache.Add(height, extCommit)
	return extCommit.Clone()
}

// LoadSeenCommit returns the locally seen Commit for the given height.
// This is useful when we've seen a commit, but there has not yet been
// a new block at `height + 1` that includes this commit in its block.LastCommit.
func (bs *DagStore) LoadSeenCommit(height int64) *types.Commit {
	comm, ok := bs.seenCommitCache.Get(height)
	if ok {
		return comm.Clone()
	}
	pbc := new(cmtproto.Commit)
	bz, err := bs.db.Get(calcSeenCommitKey(height))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = proto.Unmarshal(bz, pbc)
	if err != nil {
		panic(fmt.Sprintf("error reading block seen commit: %v", err))
	}

	commit, err := types.CommitFromProto(pbc)
	if err != nil {
		panic(fmt.Errorf("converting seen commit: %w", err))
	}
	bs.seenCommitCache.Add(height, commit)
	return commit.Clone()
}

// PruneBlocks removes block up to (but not including) a height. It returns number of blocks pruned and the evidence retain height - the height at which data needed to prove evidence must not be removed.
func (bs *DagStore) PruneBlocks(height int64, state sm.State) (uint64, int64, error) {
	if height <= 0 {
		return 0, -1, fmt.Errorf("height must be greater than 0")
	}
	bs.mtx.RLock()
	if height > bs.height {
		bs.mtx.RUnlock()
		return 0, -1, fmt.Errorf("cannot prune beyond the latest height %v", bs.height)
	}
	base := bs.base
	bs.mtx.RUnlock()
	if height < base {
		return 0, -1, fmt.Errorf("cannot prune to height %v, it is lower than base height %v",
			height, base)
	}

	pruned := uint64(0)
	batch := bs.db.NewBatch()
	defer batch.Close()
	flush := func(batch dbm.Batch, base int64) error {
		// We can't trust batches to be atomic, so update base first to make sure noone
		// tries to access missing blocks.
		bs.mtx.Lock()
		defer batch.Close()
		defer bs.mtx.Unlock()
		bs.base = base
		return bs.saveStateAndWriteDB(batch, "failed to prune")
	}

	evidencePoint := height
	for h := base; h < height; h++ {

		meta := bs.LoadBlockMeta(h)
		if meta == nil { // assume already deleted
			continue
		}

		// This logic is in place to protect data that proves malicious behavior.
		// If the height is within the evidence age, we continue to persist the header and commit data.

		if evidencePoint == height && !evidence.IsEvidenceExpired(state.LastBlockHeight, state.LastBlockTime, h, meta.Header.Time, state.ConsensusParams.Evidence) {
			evidencePoint = h
		}

		// if height is beyond the evidence point we dont delete the header
		if h < evidencePoint {
			if err := batch.Delete(calcBlockMetaKey(h)); err != nil {
				return 0, -1, err
			}
		}
		if err := batch.Delete(calcBlockHashKey(meta.BlockID.Hash)); err != nil {
			return 0, -1, err
		}
		// if height is beyond the evidence point we dont delete the commit data
		if h < evidencePoint {
			if err := batch.Delete(calcBlockCommitKey(h)); err != nil {
				return 0, -1, err
			}
		}
		if err := batch.Delete(calcSeenCommitKey(h)); err != nil {
			return 0, -1, err
		}
		for p := 0; p < int(meta.BlockID.PartSetHeader.Total); p++ {
			if err := batch.Delete(calcBlockPartKey(h, p)); err != nil {
				return 0, -1, err
			}
		}
		pruned++

		// flush every 1000 blocks to avoid batches becoming too large
		if pruned%1000 == 0 && pruned > 0 {
			err := flush(batch, h)
			if err != nil {
				return 0, -1, err
			}
			batch = bs.db.NewBatch()
			defer batch.Close()
		}
	}

	err := flush(batch, height)
	if err != nil {
		return 0, -1, err
	}
	return pruned, evidencePoint, nil
}

// SaveBlock persists the given block, blockParts, and seenCommit to the underlying db.
// blockParts: Must be parts of the block
// seenCommit: The +2/3 precommits that were seen which committed at height.
//
//	If all the nodes restart after committing a block,
//	we need this to reload the precommits to catch-up nodes to the
//	most recent height.  Otherwise they'd stall at H-1.
func (bs *DagStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	if block == nil {
		panic("DagStore can only save a non-nil block")
	}

	batch := bs.db.NewBatch()
	defer batch.Close()

	if err := bs.saveBlockToBatch(block, blockParts, seenCommit, batch); err != nil {
		panic(err)
	}

	bs.mtx.Lock()
	defer bs.mtx.Unlock()
	bs.height = block.Height
	if bs.base == 0 {
		bs.base = block.Height
	}

	// Save new DagStoreState descriptor. This also flushes the database.
	err := bs.saveStateAndWriteDB(batch, "failed to save block")
	if err != nil {
		panic(err)
	}
}

// SaveBlockWithExtendedCommit persists the given block, blockParts, and
// seenExtendedCommit to the underlying db. seenExtendedCommit is stored under
// two keys in the database: as the seenCommit and as the ExtendedCommit data for the
// height. This allows the vote extension data to be persisted for all blocks
// that are saved.
func (bs *DagStore) SaveBlockWithExtendedCommit(block *types.Block, blockParts *types.PartSet, seenExtendedCommit *types.ExtendedCommit) {
	if block == nil {
		panic("DagStore can only save a non-nil block")
	}
	if err := seenExtendedCommit.EnsureExtensions(true); err != nil {
		panic(fmt.Errorf("problems saving block with extensions: %w", err))
	}

	batch := bs.db.NewBatch()
	defer batch.Close()

	if err := bs.saveBlockToBatch(block, blockParts, seenExtendedCommit.ToCommit(), batch); err != nil {
		panic(err)
	}
	height := block.Height

	pbec := seenExtendedCommit.ToProto()
	extCommitBytes := mustEncode(pbec)
	if err := batch.Set(calcExtCommitKey(height), extCommitBytes); err != nil {
		panic(err)
	}

	bs.mtx.Lock()
	defer bs.mtx.Unlock()
	bs.height = height
	if bs.base == 0 {
		bs.base = height
	}

	// Save new DagStoreState descriptor. This also flushes the database.
	err := bs.saveStateAndWriteDB(batch, "failed to save block with extended commit")
	if err != nil {
		panic(err)
	}
}

func (bs *DagStore) saveBlockToBatch(
	block *types.Block,
	blockParts *types.PartSet,
	seenCommit *types.Commit,
	batch dbm.Batch,
) error {
	if block == nil {
		panic("DagStore can only save a non-nil block")
	}

	height := block.Height
	hash := block.Hash()

	if g, w := height, bs.Height()+1; bs.Base() > 0 && g != w {
		return fmt.Errorf("DagStore can only save contiguous blocks. Wanted %v, got %v", w, g)
	}
	if !blockParts.IsComplete() {
		return errors.New("DagStore can only save complete block part sets")
	}
	if height != seenCommit.Height {
		return fmt.Errorf("DagStore cannot save seen commit of a different height (block: %d, commit: %d)", height, seenCommit.Height)
	}

	// If the block is small, batch save the block parts. Otherwise, save the
	// parts individually.
	saveBlockPartsToBatch := blockParts.Count() <= maxBlockPartsToBatch

	// Save block parts. This must be done before the block meta, since callers
	// typically load the block meta first as an indication that the block exists
	// and then go on to load block parts - we must make sure the block is
	// complete as soon as the block meta is written.
	for i := 0; i < int(blockParts.Total()); i++ {
		part := blockParts.GetPart(i)
		bs.saveBlockPart(height, i, part, batch, saveBlockPartsToBatch)
	}

	// Save block meta
	blockMeta := types.NewBlockMeta(block, blockParts)
	pbm := blockMeta.ToProto()
	if pbm == nil {
		return errors.New("nil blockmeta")
	}
	metaBytes := mustEncode(pbm)
	if err := batch.Set(calcBlockMetaKey(height), metaBytes); err != nil {
		return err
	}
	if err := batch.Set(calcBlockHashKey(hash), []byte(fmt.Sprintf("%d", height))); err != nil {
		return err
	}

	// Save block commit (duplicate and separate from the Block)
	pbc := block.LastCommit.ToProto()
	blockCommitBytes := mustEncode(pbc)
	if err := batch.Set(calcBlockCommitKey(height-1), blockCommitBytes); err != nil {
		return err
	}

	// Save seen commit (seen +2/3 precommits for block)
	// NOTE: we can delete this at a later height
	pbsc := seenCommit.ToProto()
	seenCommitBytes := mustEncode(pbsc)
	if err := batch.Set(calcSeenCommitKey(height), seenCommitBytes); err != nil {
		return err
	}

	return nil
}

func (bs *DagStore) saveBlockPart(height int64, index int, part *types.Part, batch dbm.Batch, saveBlockPartsToBatch bool) {
	pbp, err := part.ToProto()
	if err != nil {
		panic(fmt.Errorf("unable to make part into proto: %w", err))
	}
	partBytes := mustEncode(pbp)
	if saveBlockPartsToBatch {
		err = batch.Set(calcBlockPartKey(height, index), partBytes)
	} else {
		err = bs.db.Set(calcBlockPartKey(height, index), partBytes)
	}
	if err != nil {
		panic(err)
	}
}

// Contract: the caller MUST have, at least, a read lock on `bs`.
func (bs *DagStore) saveStateAndWriteDB(batch dbm.Batch, errMsg string) error {
	bss := cmtstore.DagStoreState{
		Base:   bs.base,
		Height: bs.height,
	}
	SaveDagStoreStateBatch(&bss, batch)

	err := batch.WriteSync()
	if err != nil {
		return fmt.Errorf("error writing batch to DB %q: (base %d, height %d): %w",
			errMsg, bs.base, bs.height, err)
	}
	return nil
}

// SaveSeenCommit saves a seen commit, used by e.g. the state sync reactor when bootstrapping node.
func (bs *DagStore) SaveSeenCommit(height int64, seenCommit *types.Commit) error {
	pbc := seenCommit.ToProto()
	seenCommitBytes, err := proto.Marshal(pbc)
	if err != nil {
		return fmt.Errorf("unable to marshal commit: %w", err)
	}
	return bs.db.Set(calcSeenCommitKey(height), seenCommitBytes)
}

func (bs *DagStore) Close() error {
	return bs.db.Close()
}

//-----------------------------------------------------------------------------

func calcDagRoundKey(round int64) []byte {
	return []byte(fmt.Sprintf("D:%v", round))
}

func calcCert(hash []byte) []byte {
	return []byte(fmt.Sprintf("CT:%x", hash))
}

//-----------------------------------------------------------------------------

var dagStoreKey = []byte("dagStore")

// SaveDagStoreState persists the DagStore state to the database.
// deprecated: still present in this version for API compatibility
func SaveDagStoreState(bsj *cmtstore.DagStoreState, db dbm.DB) {
	bytes, err := proto.Marshal(bsj)
	if err != nil {
		panic(fmt.Sprintf("could not marshal state bytes: %v", err))
	}
	if db == nil {
		panic("both 'db' and 'batch' cannot be nil")
	}
	err = db.SetSync(DagStoreKey, bytes)
	if err != nil {
		panic(err)
	}

}

// LoadDagStoreState returns the DagStoreState as loaded from disk.
// If no DagStoreState was previously persisted, it returns the zero value.
func LoadDagStoreState(db dbm.DB) cmtstore.DagStoreState {
	bytes, err := db.Get(DagStoreKey)
	if err != nil {
		panic(err)
	}

	if len(bytes) == 0 {
		return cmtstore.DagStoreState{
			Base:  0,
			Round: 0,
		}
	}

	var ds cmtstore.DagStoreState
	if err := proto.Unmarshal(bytes, &ds); err != nil {
		panic(fmt.Sprintf("Could not unmarshal bytes: %X", bytes))
	}

	// Backwards compatibility with persisted data from before Base existed.
	if ds.Round > 0 && ds.Base == 0 {
		ds.Base = 1
	}
	return ds
}

//-----------------------------------------------------------------------------

// DeleteLatestBlock removes the block pointed to by height,
// lowering height by one.
func (bs *DagStore) DeleteLatestBlock() error {
	bs.mtx.RLock()
	targetHeight := bs.height
	bs.mtx.RUnlock()

	batch := bs.db.NewBatch()
	defer batch.Close()

	// delete what we can, skipping what's already missing, to ensure partial
	// blocks get deleted fully.
	if meta := bs.LoadBlockMeta(targetHeight); meta != nil {
		if err := batch.Delete(calcBlockHashKey(meta.BlockID.Hash)); err != nil {
			return err
		}
		for p := 0; p < int(meta.BlockID.PartSetHeader.Total); p++ {
			if err := batch.Delete(calcBlockPartKey(targetHeight, p)); err != nil {
				return err
			}
		}
	}
	if err := batch.Delete(calcBlockCommitKey(targetHeight)); err != nil {
		return err
	}
	if err := batch.Delete(calcSeenCommitKey(targetHeight)); err != nil {
		return err
	}
	// delete last, so as to not leave keys built on meta.BlockID dangling
	if err := batch.Delete(calcBlockMetaKey(targetHeight)); err != nil {
		return err
	}

	bs.mtx.Lock()
	defer bs.mtx.Unlock()
	bs.height = targetHeight - 1
	return bs.saveStateAndWriteDB(batch, "failed to delete the latest block")
}
