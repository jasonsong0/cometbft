package primary

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/p2p"
)

var _ p2p.Wrapper = &PrimaryDagHeader{}
var _ p2p.Wrapper = &PrimaryDagVote{}
var _ p2p.Wrapper = &PrimaryDagCert{}
var _ p2p.Unwrapper = &Message{}

// Wrap implements the p2p Wrapper interface and wraps a batch message.
func (m *PrimaryDagHeader) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_PrimaryDagHeader{PrimaryDagHeader: m}
	return mm
}

// Wrap implements the p2p Wrapper interface and wraps a batch message.
func (m *PrimaryDagVote) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_PrimaryDagVote{PrimaryDagVote: m}
	return mm
}

// Wrap implements the p2p Wrapper interface and wraps a batch message.
func (m *PrimaryDagCert) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_PrimaryDagCert{PrimaryDagCert: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped mempool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_PrimaryDagHeader:
		return m.GetPrimaryDagHeader(), nil
	case *Message_PrimaryDagVote:
		return m.GetPrimaryDagVote(), nil
	case *Message_PrimaryDagCert:
		return m.GetPrimaryDagCert(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
