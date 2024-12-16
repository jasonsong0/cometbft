package worker

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/p2p"
)

var _ p2p.Wrapper = &BatchData{}
var _ p2p.Unwrapper = &Message{}

// Wrap implements the p2p Wrapper interface and wraps a batch message.
func (m *BatchData) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_BatchData{BatchData: m}
	return mm
}

// Wrap implements the p2p Wrapper interface and wraps a batch message.
func (m *BatchAck) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_BatchAck{BatchAck: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped mempool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_BatchData:
		return m.GetBatchData(), nil
	case *Message_BatchAck:
		return m.GetBatchAck(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
