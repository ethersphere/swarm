package geth

import (
	"errors"
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
func (ps *Pss) SetPeerPublicKey(pubKeyHex string, topic string, addr string) error {
	topicFixedBytes := [4]byte{}
	topicBytes, err := hexutil.Decode(topic)
	if err != nil {
		return err
	} else if len(topicBytes) != 4 {
		return errors.New("topic must be four bytes")
	}
	copy(topicFixedBytes[:], topicBytes[:])

	addrBytes, err := hexutil.Decode(addr)
	if err != nil {
		return err
	}
	return ps.setPeerPublicKey(pubKeyHex, topicFixedBytes, addrBytes)
}

func (ps *Pss) SetPeerPublicKeyTopicBytes(pubKeyHex string, topic [4]byte, addr string) error {
	addrBytes, err := hexutil.Decode(addr)
	if err != nil {
		return err
	}
	return ps.setPeerPublicKey(pubKeyHex, topic, addrBytes)

}

func (ps *Pss) SetPeerPublicKeyAddrBytes(pubKeyHex string, topic string, addr []byte) error {
	topicFixedBytes := [4]byte{}
	topicBytes, err := hexutil.Decode(topic)
	if err != nil {
		return err
	} else if len(topicBytes) != 4 {
		return errors.New("topic must be four bytes")
	}
	copy(topicFixedBytes[:], topicBytes[:])
	return ps.setPeerPublicKey(pubKeyHex, topicFixedBytes, addr)

}

func (ps *Pss) setPeerPublicKey(pubKeyHex string, topic [4]byte, addr []byte) error {
	pubBytes, err := hexutil.Decode(pubKeyHex)
	if err != nil {
		return err
	}
	pssaddr := pss.PssAddress(addr)
	return ps.ps.SetPeerPublicKey(crypto.ToECDSAPub(pubBytes), pss.Topic{}, &pssaddr)
}
