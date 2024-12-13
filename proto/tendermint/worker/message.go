package batch

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/p2p"
)

var _ p2p.Wrapper = &Batch{}
var _ p2p.Unwrapper = &Message{}

// Wrap implements the p2p Wrapper interface and wraps a batch message.
func (m *Batch) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_Batch{Batch: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped mempool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_Batch:
		return m.GetBatch(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
