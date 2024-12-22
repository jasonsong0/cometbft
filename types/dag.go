package types

import (
	"errors"
	"fmt"
	"time"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

const ()

// DagRound will be updated when new round received to leader.
// start time is the block time of last height
type DagRound struct {
	mtx         cmtsync.Mutex
	DagRound    int64         `json:"dag_round"`
	Validators  *ValidatorSet `json:"validators"` // validator set for the dag round
	StartTime   time.Time     `json:"start_time"`
	StartHeight int64         `json:"start_height"`
	EndHeight   int64         `json:"end_height"`
	Parents     [][]byte      `json:"parents"`
}

//-------------------------------------

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
		vals, err := ValidatorSetFromProto(pb.Validators)
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

	if bm.DagRound > 1 && len(bm.Parents) == 0 {
		return fmt.Errorf("DagRound %d must have parents", bm.DagRound)
	}
	return nil
}
