package geth

import (
	"errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/swarm/pss"
)

type Pss struct {
	ps pss.Pss
}

func makeTopic(topic []byte) pss.Topic {
	topicType := pss.Topic{}
	copy(topicType[:], topic)
	return topicType
}

func (ps *Pss) SetPeerPublicKey(pubKey []byte, topic []byte, addr []byte) error {
	if len(topic) != 4 {
		return errors.New("mensch!! topic muss 4 bytes sein!")
	}
	pssaddr := pss.PssAddress(addr)
	return ps.ps.SetPeerPublicKey(crypto.ToECDSAPub(pubKey), makeTopic(topic), &pssaddr)
}

func (ps *Pss) SendAsym(pubKeyId string, topic []byte, addr []byte) error {
	pssaddr := pss.PssAddress(addr)
	return ps.ps.SendAsym(pubKeyId, makeTopic(topic), pssaddr)
}
