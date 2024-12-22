package types

import (
	"errors"
	"fmt"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
)

// BlockMeta contains meta information.
type BlockMeta struct {
	BlockID   BlockID `json:"block_id"`
	BlockSize int     `json:"block_size"`
	Header    Header  `json:"header"`
	NumTxs    int     `json:"num_txs"`
}

// NewBlockMeta returns a new BlockMeta.
func NewBlockMeta(block *Block, blockParts *PartSet) *BlockMeta {
	return &BlockMeta{
		BlockID:   BlockID{block.Hash(), blockParts.Header()},
		BlockSize: block.Size(),
		Header:    block.Header,
		NumTxs:    len(block.Data.Txs),
	}
}

func (bm *BlockMeta) ToProto() *cmtproto.BlockMeta {
	if bm == nil {
		return nil
	}

	pb := &cmtproto.BlockMeta{
		BlockID:   bm.BlockID.ToProto(),
		BlockSize: int64(bm.BlockSize),
		Header:    *bm.Header.ToProto(),
		NumTxs:    int64(bm.NumTxs),
	}
	return pb
}

func DagRoundFromProto(pb *cmtproto.DagRound) (*DagRound, error) {
	bm, err := DagRoundFromTrustedProto(pb)
	if err != nil {
		return nil, err
	}
	return bm, bm.ValidateBasic()
}

func DagRoundFromTrustedProto(pb *cmtproto.DagRound) (*DagRound, error) {
	if pb == nil {
		return nil, errors.New("daground is empty")
	}

	dr := new(DagRound)

	dr.DagRound = pb.Round
	dr.StartTime = pb.StartTime
	dr.StartHeight = pb.StartHeight
	dr.EndHeight = pb.EndHeight
	dr.Parents = pb.Parents
	if pb.Validators == nil {
		vals, err := types.ValidatorSetFromProto(pb.Validators)
		if err != nil {
			return nil, err
		}
		dr.Validators = vals
	}

	return dr, nil
}

// ValidateBasic performs basic validation.
func (bm *DagRound) ValidateBasic() error {
	if bm.DagRound <= 0 {
		return fmt.Errorf("DagRound must be greater than 0")
	}

	if bm.DagRound > 1 && len(pb.Parents) == 0 {
		return fmt.Errorf("DagRound %d must have parents", bm.DagRound)
	}
	return nil
}
