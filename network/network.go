package network

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"net"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/network/capability"
)

// BzzAddr implements the PeerAddr interface
type BzzAddr struct {
	OAddr        []byte
	UAddr        []byte
	Capabilities *capability.Capabilities
}

func (b *BzzAddr) DecodeRLP(s *rlp.Stream) error {
	_, err := s.List()
	if err != nil {
		return fmt.Errorf("list --- %v", err)
	}
	err = s.Decode(&b.OAddr)
	if err != nil {
		return fmt.Errorf("oaddr --- %v", err)
	}
	err = s.Decode(&b.UAddr)
	if err != nil {
		return fmt.Errorf("uaddr --- %v", err)
	}
	err = s.Decode(&b.Capabilities)
	if err != nil {
		return fmt.Errorf("capz --- %v", err)
	}
	return nil
}

func NewBzzAddr(oaddr []byte, uaddr []byte) *BzzAddr {
	return &BzzAddr{
		OAddr:        oaddr,
		UAddr:        uaddr,
		Capabilities: capability.NewCapabilities(),
	}
}

// EncodeRLP implements rlp.Encoder
func (b *BzzAddr) EncodeRLP(w io.Writer) error {
	err := rlp.Encode(w, b.OAddr)
	if err != nil {
		return err
	}
	err = rlp.Encode(w, b.UAddr)
	if err != nil {
		return err
	}
	y, err := rlp.EncodeToBytes(b.Capabilities)
	if err != nil {
		return err
	}
	err = rlp.Encode(w, y)
	if err != nil {
		return err
	}
	return nil
}

// DecodeRLP implements rlp.Decoder
func (b *BzzAddr) DecodeRLP(s *rlp.Stream) error {
	var err error

	b.OAddr, err = s.Bytes()
	if err != nil {
		return fmt.Errorf("oaddr --- %v", err)
	}
	b.UAddr, err = s.Bytes()
	if err != nil {
		return fmt.Errorf("uaddr --- %v", err)
	}

	y, err := s.Bytes()
	if err != nil {
		return fmt.Errorf("capsbytes --- %v", err)
	}
	err = rlp.DecodeBytes(y, &b.Capabilities)
	if err != nil {
		return fmt.Errorf("caps --- %v", err)
	}
	return nil
}

// NewBzzAddr creates a new BzzAddr with the specified byte values for over- and underlayaddresses
// It will contain an empty capabilities object
func NewBzzAddr(oaddr []byte, uaddr []byte) *BzzAddr {
	return &BzzAddr{
		OAddr:        oaddr,
		UAddr:        uaddr,
		Capabilities: capability.NewCapabilities(),
	}
}

// Address implements OverlayPeer interface to be used in Overlay.
func (a *BzzAddr) Address() []byte {
	return a.OAddr
}

// Over returns the overlay address.
func (a *BzzAddr) Over() []byte {
	return a.OAddr
}

// Under returns the underlay address.
func (a *BzzAddr) Under() []byte {
	return a.UAddr
}

// ID returns the node identifier in the underlay.
func (a *BzzAddr) ID() enode.ID {
	n, err := enode.ParseV4(string(a.UAddr))
	if err != nil {
		return enode.ID{}
	}
	return n.ID()
}

// Update updates the underlay address of a peer record
func (a *BzzAddr) Update(na *BzzAddr) *BzzAddr {
	return &BzzAddr{a.OAddr, na.UAddr, a.Capabilities}
}

// String pretty prints the address
func (a *BzzAddr) String() string {
	return fmt.Sprintf("%x <%s> cap:%s", a.OAddr, a.UAddr, a.Capabilities)
}

// RandomBzzAddr is a utility method generating a private key and corresponding enode id
// It in turn calls NewBzzAddrFromEnode to generate a corresponding overlay address from enode
func RandomBzzAddr() *BzzAddr {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("unable to generate key")
	}
	node := enode.NewV4(&key.PublicKey, net.IP{127, 0, 0, 1}, 30303, 30303)
	return NewBzzAddrFromEnode(node)
}

// NewBzzAddrFromEnode creates a BzzAddr where the overlay address is the byte representation of the enode i
// It is only used for test purposes
// TODO: This method should be replaced by (optionally deterministic) generation of addresses using NewEnode and PrivateKeyToBzzKey
func NewBzzAddrFromEnode(enod *enode.Node) *BzzAddr {
	return &BzzAddr{OAddr: enod.ID().Bytes(), UAddr: []byte(enod.URLv4()), Capabilities: capability.NewCapabilities()}
}

// WithCapabilities is a chained constructor method to set the capabilities array for a BzzAddr
func (b *BzzAddr) WithCapabilities(c *capability.Capabilities) *BzzAddr {
	b.Capabilities = c
	return b
}

// PrivateKeyToBzzKey create a swarm overlay address from the given private key
func PrivateKeyToBzzKey(prvKey *ecdsa.PrivateKey) []byte {
	pubkeyBytes := crypto.FromECDSAPub(&prvKey.PublicKey)
	return crypto.Keccak256Hash(pubkeyBytes).Bytes()
}

// EnodeParams contains the parameters used to create new Enode Records
type EnodeParams struct {
	PrivateKey *ecdsa.PrivateKey
	EnodeKey   *ecdsa.PrivateKey
	Lightnode  bool
	Bootnode   bool
}

// NewEnodeRecord creates a new valid swarm node ENR record from the given parameters
func NewEnodeRecord(params *EnodeParams) (*enr.Record, error) {

	if params.PrivateKey == nil {
		return nil, fmt.Errorf("all param private keys must be defined")
	}

	bzzkeybytes := PrivateKeyToBzzKey(params.PrivateKey)

	var record enr.Record
	record.Set(NewENRAddrEntry(bzzkeybytes))
	record.Set(ENRBootNodeEntry(params.Bootnode))
	return &record, nil
}

// NewEnode creates a new enode object for the given parameters
func NewEnode(params *EnodeParams) (*enode.Node, error) {
	record, err := NewEnodeRecord(params)
	if err != nil {
		return nil, err
	}
	err = enode.SignV4(record, params.EnodeKey)
	if err != nil {
		return nil, fmt.Errorf("ENR create fail: %v", err)
	}
	return enode.New(enode.V4ID{}, record)
}
