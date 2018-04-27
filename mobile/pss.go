package geth

import (
	"errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/swarm/pss"
)

type Pss struct {
	ps pss.Pss
}

func (ps *Pss) SetPeerPublicKey(pubKey []byte, topic []byte, addr []byte) error {
	if len(topic) != 4 {
		return errors.New("mensch!! topic muss 4 bytes sein!")
	}
	topicType := pss.Topic{}
	copy(topicType[:], topic)
	pssaddr := pss.PssAddress(addr)
	return ps.ps.SetPeerPublicKey(crypto.ToECDSAPub(pubKey), pss.Topic{}, &pssaddr)
}
