package lookup

const maxuint32 = ^uint32(0)

const lowestLevel uint8 = 0 // 0
const numLevels uint8 = 26  // 5
const HighestLevel = lowestLevel + numLevels - 1
const DefaultLevel = uint8((lowestLevel + numLevels) / 2)

const filterMask = (maxuint32 >> (32 - numLevels)) << lowestLevel

type ReadFunc func(level uint8, epoch uint32, now uint64) (interface{}, error)

type LookupParameters struct {
	Level uint8
	Epoch uint32
}

var NoClue = LookupParameters{}
var First = LookupParameters{Level: HighestLevel + 1, Epoch: 0}

func GetEpoch(t uint64, level uint8) uint32 {
	return uint32(0x00000000FFFFFFFF&t) & ((maxuint32 >> level) << level)
}

func Hint(last uint64, now uint64) LookupParameters {
	lp := LookupParameters{
		Epoch: GetEpoch(last, DefaultLevel),
		Level: DefaultLevel,
	}
	return GetNextLookup(lp, now)
}

func GetNextLevel(last LookupParameters, now uint64) uint8 {
	mix := (last.Epoch^uint32(now))&filterMask | (1 << (last.Level - 1))
	mask := uint32(1 << (lowestLevel + numLevels - 1))

	for i := uint8(lowestLevel + numLevels - 1); i > lowestLevel; i-- {
		if mix&mask != 0 {
			return i
		}
		mask = mask >> 1
	}
	return 0
}

func GetNextLookup(last LookupParameters, now uint64) LookupParameters {
	level := GetNextLevel(last, now)
	return LookupParameters{
		Level: level,
		Epoch: uint32(0x00000000FFFFFFFF&now) & ((maxuint32 >> level) << level),
	}
}

func Lookup(now uint64, hint LookupParameters, read ReadFunc) (value interface{}, err error) {
	var lastFound interface{}
	var epochMin uint32
	var epochUp uint32
	var level uint8

	if hint == NoClue {
		level = DefaultLevel
	} else {
		level = GetNextLevel(hint, now)
	}

	epoch := GetEpoch(now, level)

	for {
		if level == HighestLevel {
			epochMin = 0
		} else {
			epochMin = GetEpoch(uint64(epoch), level+1)
		}
		// try current level
		for {
			value, err = read(level, epoch, now)
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
			if epoch == epochMin {
				break
			}
			epoch -= (1 << level)
		}

		if value == nil {
			if level == HighestLevel {
				return nil, nil
			}
			if lastFound != nil {
				return lastFound, nil
			}
			level++
			epochUp = epoch
		} else {
			if epochUp == epoch {
				return value, nil
			}
			level--
			epoch += (1 << level)
		}
	}
}
