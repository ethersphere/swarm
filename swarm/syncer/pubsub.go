package syncer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/swarm/pss"
)

//
const (
	pssChunkTopic   = "SYNC" // pss topic for chunks
	pssReceiptTopic = "STOC" // pss topic for statement of custody receipts
)

// PubSub is a Postal Service interface needed to send/receive chunks, send/receive proofs
type PubSub interface {
	Register(topic string, handler func(msg []byte, p *p2p.Peer) error)
	Send(to []byte, topic string, msg []byte) error
}

// Pss implements the PubSub interface using pss
type Pss struct {
	pss  *pss.Pss // pss
	prox bool     // determines if pss send should use neighbourhood addressing
}

// NewPss creates a new Pss
func NewPss(p *pss.Pss, prox bool) *Pss {
	return &Pss{
		pss:  p,
		prox: prox,
	}
}

// Register registers a handler
func (p *Pss) Register(topic string, handler func(msg []byte, p *p2p.Peer) error) {
	f := func(msg []byte, peer *p2p.Peer, _ bool, _ string) error {
		return handler(msg, peer)
	}
	h := pss.NewHandler(f).WithRaw()
	if p.prox {
		h = h.WithProxBin()
	}
	pt := pss.BytesToTopic([]byte(topic))
	p.pss.Register(&pt, h)
}

// Send sends a message using pss SendRaw
func (p *Pss) Send(to []byte, topic string, msg []byte) error {
	pt := pss.BytesToTopic([]byte(topic))
	log.Warn("Send", "topic", topic, "to", label(to))
	return p.pss.SendRaw(pss.PssAddress(to), pt, msg)
}

// withPubSub plugs in PubSub to the storer to receive chunks and sending receipts
func (s *storer) withPubSub(ps PubSub) *storer {
	// Registers handler on pssChunkTopic that deserialises chunkMsg and calls
	// syncer's handleChunk function
	ps.Register(pssChunkTopic, func(msg []byte, p *p2p.Peer) error {
		var chmsg chunkMsg
		err := rlp.DecodeBytes(msg, &chmsg)
		if err != nil {
			return err
		}
		log.Error("Handler", "chunk", label(chmsg.Addr), "origin", label(chmsg.Origin))
		return s.handleChunk(&chmsg, p)
	})

	s.sendReceiptMsg = func(to []byte, r *receiptMsg) error {
		msg, err := rlp.EncodeToBytes(r)
		if err != nil {
			return err
		}
		log.Error("send receipt", "addr", label(r.Addr), "to", label(to))
		return ps.Send(to, pssReceiptTopic, msg)
	}
	return s
}

func label(b []byte) string {
	return hexutil.Encode(b[:2])
}

func (s *dispatcher) withPubSub(ps PubSub) *dispatcher {
	// Registers handler on pssProofTopic that deserialises proofMsg and calls
	// syncer's handleProof function
	ps.Register(pssReceiptTopic, func(msg []byte, p *p2p.Peer) error {
		var prmsg receiptMsg
		err := rlp.DecodeBytes(msg, &prmsg)
		if err != nil {
			return err
		}
		log.Error("Handler", "proof", label(prmsg.Addr), "self", label(s.baseAddr))
		return s.handleReceipt(&prmsg, p)
	})

	// consumes outgoing chunk messages and sends them to their destination
	// using neighbourhood addressing
	s.sendChunkMsg = func(c *chunkMsg) error {
		msg, err := rlp.EncodeToBytes(c)
		if err != nil {
			return err
		}
		log.Error("send chunk", "addr", label(c.Addr), "self", label(s.baseAddr))
		return ps.Send(c.Addr[:], pssChunkTopic, msg)
	}

	return s
}
