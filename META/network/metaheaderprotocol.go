package network

import (
	"encoding/binary"
)

const (
	META_CUSTOM = iota
)

const (
	_ = iota
	META_DATA_AUTHID
	META_DATA_WORK
	META_DATA_ARTIST
	META_DATA_MEDIA
	META_DATA_LICENCE
	META_DATA_USAGE
)

type METAHeader interface {
	GetUuid() uint64
	SetUuid(uint64)
	GetCommand() uint8
	SetCommand(uint8)
}

type METAEnvelope struct {
	command uint8
	uuid []byte
}

func NewMETAEnvelope() *METAEnvelope {
	return &METAEnvelope{
		command: 0,
		uuid: make([]byte, 8),
	}
}

func (mh *METAEnvelope) SetUuid(u uint64) {
	binary.LittleEndian.PutUint64(mh.uuid, uint64(u))
}

func (mh *METAEnvelope) GetUuid() uint64 {
	return binary.LittleEndian.Uint64(mh.uuid)
}

func (mh *METAEnvelope) SetCommand(u uint8) {
	mh.command = u
}

func (mh *METAEnvelope) GetCommand() uint8 {
	return mh.command
}
