package hasher

import (
	"errors"
	"sync"

	"github.com/ethersphere/swarm/file"
)

// NewAsyncWriter extends Hasher with an interface for concurrent segment/section writes
// TODO: Instead of explicitly setting double size of segment should be dynamic and chunked internally. If not, we have to keep different bmt hashers generation functions for different purposes in the same instance, or cope with added complexity of bmt hasher generation functions having to receive parameters
func (h *Hasher) NewAsyncWriter(double bool) *AsyncHasher {
	secsize := h.pool.SegmentSize
	if double {
		secsize *= 2
	}
	seccount := h.pool.SegmentCount
	if double {
		seccount /= 2
	}
	write := func(i int, section []byte, final bool) {
		h.writeSection(i, section, double, final)
	}
	return &AsyncHasher{
		Hasher:   h,
		double:   double,
		secsize:  secsize,
		seccount: seccount,
		write:    write,
		jobSize:  0,
	}
}

// AsyncHasher extends BMT Hasher with an asynchronous segment/section writer interface
// AsyncHasher cannot be used as with a hash.Hash interface: It must be used with the
// right indexes and length and the right number of sections
// It is unsafe and does not check indexes and section data lengths
//
// behaviour is undefined if
// * non-final sections are shorter or longer than secsize
// * if final section does not match length
// * write a section with index that is higher than length/secsize
// * set length in Sum call when length/secsize < maxsec
//
// * if Sum() is not called on a Hasher that is fully written
//   a process will block, can be terminated with Reset
// * it will not leak processes if not all sections are written but it blocks
//   and keeps the resource which can be released calling Reset()
type AsyncHasher struct {
	*Hasher             // extends the Hasher
	mtx      sync.Mutex // to lock the cursor access
	double   bool       // whether to use double segments (call Hasher.writeSection)
	secsize  int        // size of base section (size of hash or double)
	seccount int        // base section count
	write    func(i int, section []byte, final bool)
	errFunc  func(error)
	all      bool // if all written in one go, temporary workaround
	jobSize  int
}

// Reset implements file.SectionWriter
func (sw *AsyncHasher) Reset() {
	sw.jobSize = 0
	sw.all = false
	sw.Hasher.Reset()
}

// SetLength implements file.SectionWriter
func (sw *AsyncHasher) SetLength(length int) {
	sw.jobSize = length
}

// SetWriter implements file.SectionWriter
func (sw *AsyncHasher) SetWriter(_ file.SectionWriterFunc) file.SectionWriter {
	sw.errFunc(errors.New("Asynchasher does not currently support SectionWriter chaining"))
	return sw
}

// SectionSize implements file.SectionWriter
func (sw *AsyncHasher) SectionSize() int {
	return sw.secsize
}

// Branches implements file.SectionWriter
func (sw *AsyncHasher) Branches() int {
	return sw.seccount
}

// SeekSection locks the cursor until Write() is called; if no Write() is called, it will hang.
// Implements file.SectionWriter
func (sw *AsyncHasher) SeekSection(offset int) {
	sw.mtx.Lock()
	sw.Hasher.SeekSection(offset)
}

// Write writes to the current position cursor of the Hasher
// The cursor must first be manually set with SeekSection()
// The method will NOT advance the cursor.
// Implements file.SectionWriter
func (sw *AsyncHasher) Write(section []byte) (int, error) {
	defer sw.mtx.Unlock()
	sw.Hasher.size += len(section)
	return sw.WriteSection(sw.Hasher.cursor, section)
}

// WriteSection writes the i-th section of the BMT base
// this function can and is meant to be called concurrently
// it sets max segment threadsafely
func (sw *AsyncHasher) WriteSection(i int, section []byte) (int, error) {
	// TODO: Temporary workaround for chunkwise write
	if i < 0 {
		sw.Hasher.cursor = 0
		sw.Hasher.Reset()
		sw.Hasher.SetLength(len(section))
		sw.Hasher.Write(section)
		sw.all = true
		return len(section), nil
	}
	//sw.mtx.Lock() // this lock is now set in SeekSection
	// defer sw.mtk.Unlock() // this unlock is still left in Write()
	t := sw.getTree()
	// cursor keeps track of the rightmost section written so far
	// if index is lower than cursor then just write non-final section as is
	if i < t.cursor {
		// if index is not the rightmost, safe to write section
		go sw.write(i, section, false)
		return len(section), nil
	}
	// if there is a previous rightmost section safe to write section
	if t.offset > 0 {
		if i == t.cursor {
			// i==cursor implies cursor was set by Hash call so we can write section as final one
			// since it can be shorter, first we copy it to the padded buffer
			t.section = make([]byte, sw.secsize)
			copy(t.section, section)
			go sw.write(i, t.section, true)
			return len(section), nil
		}
		// the rightmost section just changed, so we write the previous one as non-final
		go sw.write(t.cursor, t.section, false)
	}
	// set i as the index of the righmost section written so far
	// set t.offset to cursor*secsize+1
	t.cursor = i
	t.offset = i*sw.secsize + 1
	t.section = make([]byte, sw.secsize)
	copy(t.section, section)
	return len(section), nil
}

// Sum can be called any time once the length and the span is known
// potentially even before all segments have been written
// in such cases Sum will block until all segments are present and
// the hash for the length can be calculated.
//
// b: digest is appended to b
// length: known length of the input (unsafe; undefined if out of range)
// meta: metadata to hash together with BMT root for the final digest
//   e.g., span for protection against existential forgery
//
// Implements file.SectionWriter
func (sw *AsyncHasher) Sum(b []byte) (s []byte) {
	if sw.all {
		return sw.Hasher.Sum(nil)
	}
	sw.mtx.Lock()
	t := sw.getTree()
	length := sw.jobSize
	if length == 0 {
		sw.releaseTree()
		sw.mtx.Unlock()
		s = sw.pool.zerohashes[sw.pool.Depth]
		return
	} else {
		// for non-zero input the rightmost section is written to the tree asynchronously
		// if the actual last section has been written (t.cursor == length/t.secsize)
		maxsec := (length - 1) / sw.secsize
		if t.offset > 0 {
			go sw.write(t.cursor, t.section, maxsec == t.cursor)
		}
		// set cursor to maxsec so final section is written when it arrives
		t.cursor = maxsec
		t.offset = length
		result := t.result
		sw.mtx.Unlock()
		// wait for the result or reset
		s = <-result
	}
	// relesase the tree back to the pool
	sw.releaseTree()
	meta := t.span
	// hash together meta and BMT root hash using the pools
	return doSum(sw.pool.hasher(), b, meta, s)
}
