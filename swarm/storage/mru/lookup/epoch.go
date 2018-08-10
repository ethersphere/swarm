package lookup

import "fmt"

// Epoch represents a time slot
type Epoch struct {
	Level    uint8
	BaseTime uint64
}

const EpochLength = 1 + 8

func (e *Epoch) LaterThan(epoch Epoch) bool {
	if e.BaseTime == epoch.BaseTime {
		return e.Level < epoch.Level
	}
	return e.BaseTime >= epoch.BaseTime
}

func (e *Epoch) String() string {
	return fmt.Sprintf("Epoch{BaseTime:%d, Level:%d", e.BaseTime, e.Level)
}
