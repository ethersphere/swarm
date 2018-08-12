package lookup

const maxuint64 = ^uint64(0)

const lowestLevel uint8 = 0 // 0
const numLevels uint8 = 26  //5
const highestLevel = lowestLevel + numLevels - 1
const defaultLevel = highestLevel

const filterMask = (maxuint64 >> (64 - numLevels)) << lowestLevel

// ReadFunc is a handler called by Lookup each time it attempts to find a value
// It should return <nil> if a value is not found
// It should return <nil> if a value is found, but its timestamp is higher than "now"
// It should only return an error in case the handler wants to stop the
// lookup process entirely.
type ReadFunc func(epoch Epoch, now uint64) (interface{}, error)

// Hint that can be provided when the Lookup caller does not have
// a clue about where the last update may be
var NoClue = Epoch{}

func getBaseTime(t uint64, level uint8) uint64 {
	return t & ((maxuint64 >> level) << level)
}

func Hint(last uint64) Epoch {
	return Epoch{
		Time:  last,
		Level: defaultLevel,
	}
}

func getNextLevel(last Epoch, now uint64) uint8 {
	mix := (last.Base() ^ now) | (1 << (last.Level - 1))
	if mix > (maxuint64 >> (64 - highestLevel - 1)) {
		return highestLevel
	}
	mask := uint64(1 << (highestLevel))

	for i := uint8(highestLevel); i > lowestLevel; i-- {
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
	if last == NoClue {
		return GetFirstEpoch(now)
	}
	level := getNextLevel(last, now)
	return Epoch{
		Level: level,
		Time:  now,
	}
}

// GetFirstEpoch returns the epoch where the first update should be located
// based on what time it is now.
func GetFirstEpoch(now uint64) Epoch {
	return Epoch{Level: highestLevel, Time: now}
}

// Lookup finds the update with the highest timestamp that is smaller or equal than 'now'
// It takes a hint which should be the epoch where the last known update was
// If you don't know in what epoch the last update happened, simply submit lookup.NoClue
// read() will be called on each lookup attempt
// Returns an error only if read() returns an error
// Returns nil if an update was not found
func Lookup(now uint64, hint Epoch, read ReadFunc) (value interface{}, err error) {
	var lastFound interface{}
	var baseTimeMin uint64
	var baseTimeUp = maxuint64
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
			baseTimeMin = getBaseTime(baseTime, level+1)
		}
		// try current level
		for {
			value, err = read(Epoch{Level: level, Time: baseTime}, now)
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
