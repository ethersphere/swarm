package testutil

// the actual data length generated (could be longer than max datalength of the BMT)
const BufferSize = 4128

const (
	// segmentCount is the maximum number of segments of the underlying chunk
	// Should be equal to max-chunk-data-size / hash-size
	// Currently set to 128 == 4096 (default chunk size) / 32 (sha3.keccak256 size)
	SegmentCount = 128
)

var Counts = []int{1, 2, 3, 4, 5, 8, 9, 15, 16, 17, 32, 37, 42, 53, 63, 64, 65, 111, 127, 128}
