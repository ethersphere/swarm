package lookup

const maxuint32 = ^uint32(0)

const lowestLevel uint8 = 0 // 0
const numLevels uint8 = 26  // 5
const highestLevel = lowestLevel + numLevels - 1
const defaultLevel = uint8((lowestLevel + numLevels) / 2)

const filterMask = (maxuint32 >> (32 - numLevels)) << lowestLevel

// ReadFunc is a handler called by Lookup each time it attempts to find a value
// It should return <nil> if a value is not found
// It should return <nil> if a value is found, but its timestamp is higher than "now"
// It should only return an error in case the handler wants to stop the
// lookup process entirely.
type ReadFunc func(epoch Epoch, now uint64) (interface{}, error)

// Epoch represents a time slot
type Epoch struct {
	Level    uint8
	BaseTime uint32
}

// Hint that can be provided when the Lookup caller does not have
// a clue about where the last update may be
var NoClue = Epoch{}

func getBaseTime(t uint64, level uint8) uint32 {
	return uint32(0x00000000FFFFFFFF&t) & ((maxuint32 >> level) << level)
}

func Hint(last uint64, now uint64) Epoch {
	lp := Epoch{
		BaseTime: getBaseTime(last, defaultLevel),
		Level:    defaultLevel,
	}
	return GetNextEpoch(lp, now)
}

func getNextLevel(last Epoch, now uint64) uint8 {
	mix := (last.BaseTime^uint32(now))&filterMask | (1 << (last.Level - 1))
	mask := uint32(1 << (lowestLevel + numLevels - 1))

	for i := uint8(lowestLevel + numLevels - 1); i > lowestLevel; i-- {
		if mix&mask != 0 {
			return i
		}
		mask = mask >> 1
	}
	return 0
}

// GetNextEpoch returns the epoch where the next update should be located
// according to where the previous update was
// and what time it is now.
func GetNextEpoch(last Epoch, now uint64) Epoch {
	level := getNextLevel(last, now)
	return Epoch{
		Level:    level,
		BaseTime: uint32(0x00000000FFFFFFFF&now) & ((maxuint32 >> level) << level),
	}
}

// GetFirstEpoch returns the epoch where the first update should be located
// and what time it is now.
func GetFirstEpoch(now uint64) Epoch {
	return GetNextEpoch(Epoch{Level: highestLevel + 1, BaseTime: 0}, now)
}

// Lookup finds the update with the highest timestamp that is smaller or equal than 'now'
// It takes a hint which should be the epoch where the last known update was
// If you don't know in what epoch the last update happened, simply submit lookup.NoClue
// read will be called on each lookup attempt
// Returns an error only if read returns an error
// Returns nil if an update was not found
func Lookup(now uint64, hint Epoch, read ReadFunc) (value interface{}, err error) {
	var lastFound interface{}
	var baseTimeMin uint32
	var baseTimeUp uint32
	var level uint8

	if hint == NoClue {
		level = defaultLevel
	} else {
		level = getNextLevel(hint, now)
	}

	baseTime := getBaseTime(now, level)

	for {
		if level == highestLevel {
			baseTimeMin = 0
		} else {
			baseTimeMin = getBaseTime(uint64(baseTime), level+1)
		}
		// try current level
		for {
			value, err = read(Epoch{Level: level, BaseTime: baseTime}, now)
			if err != nil {
				return nil, err
			}
			if value != nil {
				lastFound = value
				if level == lowestLevel {
					return value, nil
				}
				break
			}
			if baseTime == baseTimeMin {
				break
			}
			baseTime -= (1 << level)
		}

		if value == nil {
			if level == highestLevel {
				return nil, nil
			}
			if lastFound != nil {
				return lastFound, nil
			}
			level++
			baseTimeUp = baseTime
		} else {
			if baseTimeUp == baseTime {
				return value, nil
			}
			level--
			baseTime += (1 << level)
		}
	}
}
