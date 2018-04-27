package geth

import (
	//"crypto/ecdsa"
	//	"github.com/ethereum/go-ethereum/common"
	//	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/pss"
)

type Pss struct {
	ps *pss.Pss
}

func (ps *Pss) SetPeerPublicKey(pubkeyHex string, topic [4]byte, addr []byte) error {
	//privKey, err := HexToECDSA(pubkeyHex)
	//return ps.SetPeerPublicKey(privKey
	return nil
}
