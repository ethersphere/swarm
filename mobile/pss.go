package geth

import (
	//"crypto/ecdsa"
	//	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/pss"
)

type Pss struct {
	ps pss.Pss
}

//func (ps *Pss) SetPeerPublicKey(pubKeyHex string, topic [4]byte, addr []byte) error {
func (ps *Pss) SetPeerPublicKey(pubKeyHex string) error {
	pubBytes, err := hexutil.Decode(pubKeyHex)
	if err != nil {
		return err
	}
	addr := []byte{}
	pssaddr := pss.PssAddress(addr)
	//return ps.ps.SetPeerPublicKey(crypto.ToECDSAPub(pubBytes), topic, &pssaddr)
	return ps.ps.SetPeerPublicKey(crypto.ToECDSAPub(pubBytes), pss.Topic{}, &pssaddr)
}
