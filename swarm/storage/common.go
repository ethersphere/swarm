package storage

import (
	"context"
	"time"
)

var putTimeout = 30 * time.Second

// PutChunks adds chunks  to localstore
// It waits for receive on the stored channel
// It logs but does not fail on delivery error
func PutChunks(store *LocalStore, chunks ...Chunk) error {
	i := 0
	f := func(n int64) Chunk {
		chunk := chunks[i]
		i++
		return chunk
	}
	_, err := put(store, len(chunks), f)
	return err
}

func put(store *LocalStore, n int, f func(i int64) Chunk) (hs []Address, err error) {
	// put to localstore and wait for stored channel
	// does not check delivery error state
	done := make(chan struct{})
	errc := make(chan error)
	ctx, _ := context.WithTimeout(context.Background(), putTimeout)
	// defer cancel()
	defer close(done)
	for i := int64(0); i < int64(n); i++ {
		chunk := f(DefaultChunkSize)
		wait, err := store.Put(chunk)
		if err != nil {
			return nil, err
		}
		go func() {
			select {
			case errc <- wait(ctx):
			case <-done:
			}
		}()
		hs = append(hs, chunk.Address())
	}

	// wait for all chunks to be stored
	for i := 0; i < n; i++ {
		err := <-errc
		if err != nil {
			return nil, err
		}
	}
	return hs, nil
}
