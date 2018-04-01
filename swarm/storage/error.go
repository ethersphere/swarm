package storage

const (
	ErrInit = iota
	ErrNotFound
	ErrIO
	ErrUnauthorized
	ErrInvalidValue
	ErrDataOverflow
	ErrNothingToReturn
	ErrInvalidSignature
	ErrNotSynced
	ErrPeriodDepth
	ErrCnt
)

const (
	ChunkErrOk = iota
	ChunkErrNotFound
	ChunkErrNoForward
	ChunkErrTimeout
	ChunkErrInvalid
)

type ChunkError byte
