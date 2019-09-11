package ttlset_test

import (
	"testing"
	"time"

	"github.com/epiclabs-io/ut"
	"github.com/ethersphere/swarm/pss/internal/ttlset"
	"github.com/tilinna/clock"
)

func TestTTLSet(tx *testing.T) {
	t := ut.BeginTest(tx, false) // set to true to generate test results
	defer t.FinishTest()
	var err error

	testClock := clock.NewMock(time.Unix(0, 0))
	waitClean := make(chan bool)

	testEntryTTL := 10 * time.Second
	testSet := ttlset.New(&ttlset.Config{
		EntryTTL: testEntryTTL,
		Clock:    testClock,
		OnClean: func() {
			waitClean <- true
		},
	})

	key1 := "some key"
	key2 := "some other key"

	// check adding a key to the set
	err = testSet.Add(key1)
	t.Ok(err)

	// check if the key is now there:
	hasKey := testSet.Has(key1)
	t.Assert(hasKey == true, "key1 should've been in the set, but Has() returned false")

	// check if Has() returns false when asked about a key that was never added:
	hasKey = testSet.Has("some made up key")
	t.Assert(hasKey == false, "Has() should have returned false when presented with a key that was never added")

	// Let some time pass, but not enough to have the key expire:
	testClock.Add(testEntryTTL / 2)

	// check if the key is still there:
	hasKey = testSet.Has(key1)
	t.Assert(hasKey == true, "key1 should've been in the set, but Has() returned false")

	// Let some time pass well beyond the expiry time, so key1 expires:
	testClock.Add(testEntryTTL * 2)

	<-waitClean // Will only continue if the clean function was indeed called

	// Add another key to the set:
	err = testSet.Add(key2)
	t.Ok(err)

	hasKey = testSet.Has(key1)
	t.Assert(hasKey == false, "key1 should've been removed from the set, but Has() returned true")

	hasKey = testSet.Has(key2)
	t.Assert(hasKey == true, "key should remain in the set, but Has() returned false")

	// Let some time pass well beyond key2's expiry time, so key2 expires:
	testClock.Add(testEntryTTL * 2)
	<-waitClean // Will only continue if the clean function was indeed called

	hasKey = testSet.Has(key2)
	t.Assert(hasKey == false, "key2 should have been wiped, but Has() returned true")

	// stop the service
	err = testSet.Stop()
	t.Ok(err)

	// stopping again must return an error
	err = testSet.Stop()
	t.MustFail(err, "Expected Stop() to fail if the service is already stopped")
}
