package hasher

import (
	"fmt"
	"sync"
)

// keeps an index of all the existing jobs for a file hashing operation
// sorted by level
//
// it also keeps all the "top hashes", ie hashes on first data section index of every level
// these are needed in case of balanced tree results, since the hashing result would be
// lost otherwise, due to the job not having any intermediate storage of any data
type jobIndex struct {
	maxLevels int
	jobs      []sync.Map
	topHashes [][]byte
	mu        sync.Mutex
}

func newJobIndex(maxLevels int) *jobIndex {
	ji := &jobIndex{
		maxLevels: maxLevels,
	}
	for i := 0; i < maxLevels; i++ {
		ji.jobs = append(ji.jobs, sync.Map{})
	}
	return ji
}

// implements Stringer interface
func (ji *jobIndex) String() string {
	return fmt.Sprintf("%p", ji)
}

// Add adds a job to the index at the level
// and data section index specified in the job
func (ji *jobIndex) Add(jb *job) {
	//log.Trace("adding job", "job", jb)
	ji.jobs[jb.level].Store(jb.dataSection, jb)
}

// Get retrieves a job from the job index
// based on the level of the job and its data section index
// if a job for the level and section index does not exist this method returns nil
func (ji *jobIndex) Get(lvl int, section int) *job {
	jb, ok := ji.jobs[lvl].Load(section)
	if !ok {
		return nil
	}
	return jb.(*job)
}

// Delete removes a job from the job index
// leaving it to be garbage collected when
// the reference in the main code is relinquished
func (ji *jobIndex) Delete(jb *job) {
	ji.jobs[jb.level].Delete(jb.dataSection)
}

// AddTopHash should be called by a job when a hash is written to the first index of a level
// since the job doesn't store any data written to it (just passing it through to the underlying writer)
// this is needed for the edge case of balanced trees
func (ji *jobIndex) AddTopHash(ref []byte) {
	ji.mu.Lock()
	defer ji.mu.Unlock()
	ji.topHashes = append(ji.topHashes, ref)
	//log.Trace("added top hash", "length", len(ji.topHashes), "index", ji)
}

// GetJobHash gets the current top hash for a particular level set by AddTopHash
func (ji *jobIndex) GetTopHash(lvl int) []byte {
	ji.mu.Lock()
	defer ji.mu.Unlock()
	return ji.topHashes[lvl-1]
}

func (ji *jobIndex) GetTopHashLevel() int {
	ji.mu.Lock()
	defer ji.mu.Unlock()
	return len(ji.topHashes)
}
