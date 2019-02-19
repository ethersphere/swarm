package constants

const (
	DefaultLDBCapacity                = 5000000 // capacity for LevelDB, by default 5*10^6*4096 bytes == 20GB
	DefaultCacheCapacity              = 10000   // capacity for in-memory chunks' cache
	DefaultChunkRequestsCacheCapacity = 5000000 // capacity for container holding outgoing requests for chunks. should be set to LevelDB capacity
	OpenFileLimit                     = 128
	MaxPO                             = 16
	AddressLength                     = 32
	DefaultGCRatio                    = 10
	DefaultMaxGCRound                 = 10000
	DefaultMaxGCBatch                 = 5000

	DefaultChunkSize = 4096

	WwEntryCnt  = 1 << 0
	WwIndexCnt  = 1 << 1
	WwAccessCnt = 1 << 2
)
