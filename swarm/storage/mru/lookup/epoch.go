package lookup

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Epoch represents a time slot
type Epoch struct {
	Level uint8  `json: "level"`
	Time  uint64 `json: "time"`
}

type EpochID [8]byte

const EpochLength = 8
const MaxTime uint64 = (1 << 56) - 1

func (e *Epoch) Base() uint64 {
	return getBaseTime(e.Time, e.Level)
}

func (e *Epoch) ID() EpochID {
	base := e.Base()
	var id EpochID
	binary.LittleEndian.PutUint64(id[:], base)
	id[7] = e.Level
	return id
}

// MarshalBinary implements the encoding.BinaryMarshaller interface
func (e *Epoch) MarshalBinary() (data []byte, err error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b[:], e.Time)
	b[7] = e.Level
	return b, nil
}

// MarshalBinary implements the encoding.BinaryUnmarshaller interface
func (e *Epoch) UnmarshalBinary(data []byte) error {
	if len(data) != EpochLength {
		return errors.New("Invalid data unmarshalling Epoch")
	}
	b := make([]byte, 8)
	copy(b, data)
	e.Level = b[7]
	b[7] = 0
	e.Time = binary.LittleEndian.Uint64(b)
	return nil
}

func (e *Epoch) LaterThan(epoch Epoch) bool {
	if e.Time == epoch.Time {
		return e.Level < epoch.Level
	}
	return e.Time >= epoch.Time
}

func (e *Epoch) Equals(epoch Epoch) bool {
	return e.Level == epoch.Level && e.Base() == epoch.Base()
}

func (e *Epoch) String() string {
	return fmt.Sprintf("Epoch{Time:%d, Level:%d}", e.Time, e.Level)
}
