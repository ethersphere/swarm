package ttlset_test

import (
	"testing"
	"time"

	"github.com/ethersphere/swarm/pss/internal/ttlset"
	"github.com/tilinna/clock"
)

func TestTTLSet(t *testing.T) {
	var err error

	testClock := clock.NewMock(time.Unix(0, 0))

	testEntryTTL := 10 * time.Second
	testSet := ttlset.New(&ttlset.Config{
		EntryTTL: testEntryTTL,
		Clock:    testClock,
	})

	key1 := "some key"
	key2 := "some other key"

	// check adding a key to the set
	err = testSet.Add(key1)
	if err != nil {
		t.Fatal((err))
	}

	// check if the key is now there:
	hasKey := testSet.Has(key1)
	if !(hasKey == true) {
		t.Fatal("key1 should've been in the set, but Has() returned false")
	}

	// check if Has() returns false when asked about a key that was never added:
	hasKey = testSet.Has("some made up key")
	if !(hasKey == false) {
		t.Fatal("Has() should have returned false when presented with a key that was never added")
	}

	// Let some time pass, but not enough to have the key expire:
	testClock.Add(testEntryTTL / 2)

	// check if the key is still there:
	hasKey = testSet.Has(key1)
	if !(hasKey == true) {
		t.Fatal("key1 should've been in the set, but Has() returned false")
	}

	// Let some time pass well beyond the expiry time, so key1 expires:
	testClock.Add(testEntryTTL * 2)

	// Add another key to the set:
	err = testSet.Add(key2)
	if err != nil {
		t.Fatal((err))
	}

	hasKey = testSet.Has(key1)
	if !(hasKey == false) {
		t.Fatal("key1 should've been removed from the set, but Has() returned true")
	}

	hasKey = testSet.Has(key2)
	if !(hasKey == true) {
		t.Fatal("key should remain in the set, but Has() returned false")
	}

	// Let some time pass well beyond key2's expiry time, so key2 expires:
	testClock.Add(testEntryTTL * 2)

	hasKey = testSet.Has(key2)
	if !(hasKey == false) {
		t.Fatal("key2 should have been wiped, but Has() returned true")
	}
}

func TestGC(t *testing.T) {
	var err error

	testClock := clock.NewMock(time.Unix(0, 0))

	testEntryTTL := 10 * time.Second
	testSet := ttlset.New(&ttlset.Config{
		EntryTTL: testEntryTTL,
		Clock:    testClock,
	})

	key1 := "some key"
	key2 := "some later key"

	// check adding a message to the cache
	err = testSet.Add(key1)
	if err != nil {
		t.Fatal((err))
	}

	// move the clock 2 seconds
	testClock.Add(2 * time.Second)

	// add a second key which will have a later expiration time
	err = testSet.Add(key2)
	if err != nil {
		t.Fatal((err))
	}

	count := testSet.Count()
	if !(count == 2) {
		t.Fatal("Expected the set to contain 2 keys")
	}

	testSet.GC() // attempt a cleanup. This cleanup should not affect any of the two keys, since they are not expired.

	count = testSet.Count()
	if !(count == 2) {
		t.Fatal("Expected the set to still contain 2 keys")
	}

	//Now, move the clock forward 9 seconds. This will expire key1 but still keep key2
	testClock.Add(9 * time.Second)
	testSet.GC() // invoke the internal cleaning function, which should wipe only key1
	count = testSet.Count()
	if !(count == 1) {
		t.Fatal("Expected the set to now have only 1 key")
	}

	//Verify if key1 was wiped but key2 persists:
	hasKey := testSet.Has(key1)
	if !(hasKey == false) {
		t.Fatal("Expected the set to have removed key1")
	}

	hasKey = testSet.Has(key2)
	if !(hasKey == true) {
		t.Fatal("Expected the set to still contain key2")
	}

	//Now, move the clock some more time. This will wipe key2
	testClock.Add(7 * time.Second)
	testSet.GC() // invoke the internal cleaning function, which should wipe only key1

	count = testSet.Count()
	// verify the map is now empty
	if !(count == 0) {
		t.Fatal("Expected the set to be empty")
	}

}
