package network

import (
	"crypto/ecdsa"
	"fmt"
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

// RandomBzzAddr is a utility method generating an address from a public key
func RandomBzzAddr() *BzzAddr {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("unable to generate key")
	}
	node := enode.NewV4(&key.PublicKey, net.IP{127, 0, 0, 1}, 30303, 30303)
	return NewBzzAddrFromEnode(node)
}

func NewBzzAddrFromEnode(enod *enode.Node) *BzzAddr {
	return &BzzAddr{OAddr: enod.ID().Bytes(), UAddr: []byte(enod.URLv4()), Capabilities: capability.NewCapabilities()}
}

func (b *BzzAddr) WithCapabilities(c *capability.Capabilities) *BzzAddr {
	b.Capabilities = c
	return b
}

func PrivateKeyToBzzKey(prvKey *ecdsa.PrivateKey) []byte {
	pubkeyBytes := crypto.FromECDSAPub(&prvKey.PublicKey)
	return crypto.Keccak256Hash(pubkeyBytes).Bytes()
}

type EnodeParams struct {
	PrivateKey *ecdsa.PrivateKey
	EnodeKey   *ecdsa.PrivateKey
	Lightnode  bool
	Bootnode   bool
}

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
