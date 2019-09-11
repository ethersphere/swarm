package ttlset

import (
	"testing"
	"time"

	"github.com/epiclabs-io/ut"
	"github.com/tilinna/clock"
)

// white-box testing for automatic cleaning feature

func TestClean(tx *testing.T) {
	t := ut.BeginTest(tx, false) // set to true to generate test results
	defer t.FinishTest()
	var err error

	testClock := clock.NewMock(time.Unix(0, 0))

	testEntryTTL := 10 * time.Second
	testSet := New(&Config{
		EntryTTL: testEntryTTL,
		Clock:    testClock,
	})

	key1 := "some key"
	key2 := "some later key"

	// check adding a message to the cache
	err = testSet.Add(key1)
	t.Ok(err)

	// move the clock 2 seconds
	testClock.Add(2 * time.Second)

	// add a second key which will have a later expiration time
	err = testSet.Add(key2)
	t.Ok(err)

	// Check if both keys were added to the internal map.
	_, hasKey := testSet.set[key1]
	t.Assert(hasKey == true, "Expected the set to contain key1")
	_, hasKey = testSet.set[key2]
	t.Assert(hasKey == true, "Expected the set to contain key2")

	testSet.clean() // attempt a cleanup. This cleanup should not affect any of the two keys, since they are not expired.

	// Thus, check if both keys are still in the internal map:
	_, hasKey = testSet.set[key1]
	t.Assert(hasKey == true, "Expected the set to still contain key1")
	_, hasKey = testSet.set[key2]
	t.Assert(hasKey == true, "Expected the set to still contain key2")

	//Now, move the clock forward 9 seconds. This will have the effect of wiping key1 but keeping key2
	testClock.Add(9 * time.Second)
	testSet.clean() // invoke the internal cleaning function, which should wipe only key1

	//Verify if key1 was wiped but key2 persists:
	_, hasKey = testSet.set[key1]
	t.Assert(hasKey == false, "Expected the set to have removed key1")
	_, hasKey = testSet.set[key2]
	t.Assert(hasKey == true, "Expected the set to still contain key2")

	//Now, move the clock some more time. This will wipe key2
	testClock.Add(7 * time.Second)
	testSet.clean() // invoke the internal cleaning function, which should wipe only key1

	// verify the map is now empty
	t.Assert(len(testSet.set) == 0, "Expected the set to be empty")

}
