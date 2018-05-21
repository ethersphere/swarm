package storage

import (
	"sync"
)

// put to localstore and wait for stored channel
// does not check delivery error state
func PutChunks(store *LocalStore, chunks ...*Chunk) {
	wg := sync.WaitGroup{}
	wg.Add(len(chunks))
	go func() {
		for _, c := range chunks {
			<-c.dbStoredC
			wg.Done()
		}
	}()
	for _, c := range chunks {
		go store.Put(c)
	}
	wg.Wait()
}
