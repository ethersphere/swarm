// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

/*
Package lookup defines resource lookup algorithms and provides tools to place updates
so they can be found
*/
package lookup

const maxuint64 = ^uint64(0)
const lowestLevel uint8 = 0
const numLevels uint8 = 26
const highestLevel = lowestLevel + numLevels - 1
const defaultLevel = highestLevel

// ReadFunc is a handler called by Lookup each time it attempts to find a value
// It should return <nil> if a value is not found
// It should return <nil> if a value is found, but its timestamp is higher than "now"
// It should only return an error in case the handler wants to stop the
// lookup process entirely.
type ReadFunc func(epoch Epoch, now uint64) (interface{}, error)

// NoClue is a hint that can be provided when the Lookup caller does not have
// a clue about where the last update may be
var NoClue = Epoch{}

func getBaseTime(t uint64, level uint8) uint64 {
	return t & (maxuint64 << level)
}

// Hint creates a hint based only on the last known update time
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
	var epoch Epoch
	if hint == NoClue {
		hint = Epoch{Time: 0, Level: highestLevel}
	}

	t := now

	for {
		epoch = GetNextEpoch(hint, t)
		value, err = read(epoch, now)
		if err != nil {
			return nil, err
		}
		if value != nil {
			lastFound = value
			if epoch.Level == lowestLevel || epoch.Equals(hint) {
				return value, nil
			}
			hint = epoch
		} else {
			if epoch.Base() == hint.Base() {
				if lastFound != nil {
					return lastFound, nil
				}
				// we have reached the hint itself
				// check it out
				value, err = read(hint, now)
				if err != nil {
					return nil, err
				}
				if value != nil {
					return value, nil
				}
				// bad hint.
				epoch = hint
				hint = Epoch{Time: 0, Level: highestLevel}
			}
			base := epoch.Base()
			if base == 0 {
				return nil, nil
			}
			t = base - 1
		}
	}
}

// Lookup2 is a slower alternative lookup algorithm
func Lookup2(now uint64, hint Epoch, read ReadFunc) (value interface{}, err error) {
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
