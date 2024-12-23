package types

import (
	"crypto/sha256"
	"time"

	primaryproto "github.com/cometbft/cometbft/proto/tendermint/primary"
)

// hash for dag header
const DagHeaderKeySize = sha256.Size
const DagCertKeySize = sha256.Size

type DagHeaderKey [DagHeaderKeySize]byte
type DagCertKey [DagCertKeySize]byte

type DagHeaderProc struct {
	Hdr primaryproto.PrimaryDagHeader

	Votes      map[string]*primaryproto.DagVote
	VoteWeight map[string]float64 // acked and voting power

	Certificate primaryproto.PrimaryDagCert

	PrimaryCreated     time.Time
	CertificateCreated time.Time
}

func (d *DagHeaderKey) ToBytes() []byte {
	b := make([]byte, DagHeaderKeySize)
	copy(b, d[:])
	return b
}

func (d *DagCertKey) ToBytes() []byte {
	b := make([]byte, DagCertKeySize)
	copy(b, d[:])
	return b
}

func GenDagHeaderKey(header *primaryproto.PrimaryDagHeader) (pk DagHeaderKey, err error) {

	data, err := header.Marshal()
	if err != nil {
		return
	}

	s := sha256.New()
	_, err = s.Write(data)
	if err != nil {
		return
	}

	copy(pk[:], s.Sum(nil))
	return
}

func GenCertDigest(cert *primaryproto.PrimaryDagCert) (digest DagCertKey, err error) {
	data, err := cert.Marshal()
	if err != nil {
		return
	}

	s := sha256.New()
	_, err = s.Write(data)
	if err != nil {
		return
	}

	copy(digest[:], s.Sum(nil))
	return

}

/*
func ComputeProtoSizeForBatches(batches []Batch) int {
	sum := int(0)
	for _, b := range batches {
		p := b.ToProto()
		sum = sum + p.Size()
	}
	return sum
}
*/
