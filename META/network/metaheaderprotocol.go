package network

import (
	"encoding/binary"
)

const (
	META_CUSTOM = iota
)

type METAHeader interface {
	GetUuid() uint64
	SetUuid(uint64)
	GetCommand() uint8
	SetCommand(uint8)
}

type METAEnvelope struct {
	Command uint8
	Uuid []byte
}

func NewMETAEnvelope() *METAEnvelope {
	return &METAEnvelope{
		Command: 0,
		Uuid: make([]byte, 8),
	}
}

func (mh *METAEnvelope) SetUuid(u uint64) {
	binary.LittleEndian.PutUint64(mh.Uuid, uint64(u))
}

func (mh *METAEnvelope) GetUuid() uint64 {
	return binary.LittleEndian.Uint64(mh.Uuid)
}

func (mh *METAEnvelope) SetCommand(u uint8) {
	mh.Command = u
}

func (mh *METAEnvelope) GetCommand() uint8 {
	return mh.Command
}
