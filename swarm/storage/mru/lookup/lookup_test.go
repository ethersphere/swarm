package lookup_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"
)

type Data struct {
	Payload uint64
	Time    uint64
}

type Store map[lookup.Epoch]*Data

func write(store Store, epoch lookup.Epoch, value *Data) {
	fmt.Printf("Write: %d-%d\n", epoch.BaseTime, epoch.Level)
	store[epoch] = value
}

//var last uint64
//var lastLevel = lookup.HighestLevel + 1

func update(store Store, last lookup.Epoch, now uint64, value *Data) lookup.Epoch {
	var epoch lookup.Epoch

	if last == lookup.NoClue {
		epoch = lookup.GetFirstEpoch(now)
	} else {
		epoch = lookup.GetNextEpoch(last, now)
	}

	write(store, epoch, value)

	return epoch
}

const Day = 60 * 60 * 24
const Year = Day * 365
const Month = Day * 30

func makeReadFunc(store Store, counter *int) lookup.ReadFunc {
	return func(epoch lookup.Epoch, now uint64) (interface{}, error) {
		fmt.Printf("Read: %d-%d\n", epoch.BaseTime, epoch.Level)
		*counter++
		data := store[epoch]
		if data != nil && data.Time <= now {
			return data, nil
		}
		return nil, nil
	}
}

func TestLookup(t *testing.T) {

	store := make(Store)
	readCount := 0
	readFunc := makeReadFunc(store, &readCount)

	// write an update every month for 12 months 3 years ago and then silence for two years

	now := uint64(1533799046)
	epoch := lookup.NoClue

	var lastData *Data
	for i := uint64(0); i < 12; i++ {
		t := uint64(now - Year*3 + i*Month) // update every month for 12 months 3 years ago and then silence for two years
		data := Data{
			Payload: t, //our "payload" will be the timestamp itself.
			Time:    t,
		}
		epoch = update(store, epoch, t, &data)
		lastData = &data
	}

	// try to get the last value

	value, err := lookup.Lookup(now, lookup.NoClue, readFunc)
	if err != nil {
		t.Fatal(err)
	}

	readCountWithoutHint := readCount

	if value != lastData {
		t.Fatalf("Expected lookup to return the last written value: %v. Got %v", lastData, value)
	}

	// reset the read count for the next test
	readCount = 0
	// Provide a hint to get a faster lookup. In particular, we give the exact location of the last update
	value, err = lookup.Lookup(now, epoch, readFunc)
	if err != nil {
		t.Fatal(err)
	}

	if value != lastData {
		t.Fatalf("Expected lookup to return the last written value: %v. Got %v", lastData, value)
	}

	if readCount > readCountWithoutHint {
		t.Fatalf("Expected lookup to complete with fewer reads than %d since we provided a hint. Did %d reads.", readCountWithoutHint, readCount)
	}

	// try to get an intermediate value
	// if we look for a value in now - Year*3 + 6*Month, we should get that value
	// Since the "payload" is the timestamp itself, we can check this.

	expectedTime := now - Year*3 + 6*Month

	value, err = lookup.Lookup(expectedTime, lookup.NoClue, readFunc)
	if err != nil {
		t.Fatal(err)
	}

	data, ok := value.(*Data)

	if !ok {
		t.Fatal("Expected value to contain data")
	}

	if data.Time != expectedTime {
		t.Fatalf("Expected value timestamp to be %d, got %d", data.Time, expectedTime)
	}

}

func TestLookupFail(t *testing.T) {

	store := make(Store)
	readCount := 0

	readFunc := makeReadFunc(store, &readCount)
	now := uint64(1533799046)

	// don't write anything and try to look up.
	// we're testing we don't get stuck in a loop

	value, err := lookup.Lookup(now, lookup.NoClue, readFunc)
	if err != nil {
		t.Fatal(err)
	}
	if value != nil {
		t.Fatal("Expected value to be nil, since the update should've failed")
	}

	// write an update every second for the last 1000 seconds

	epoch := lookup.NoClue

	var lastData *Data
	for i := uint64(0); i <= 1000; i++ {
		t := uint64(now - 1000 + i) // update every second for the last 1000 seconds
		data := Data{
			Payload: t, //our "payload" will be the timestamp itself.
			Time:    t,
		}
		epoch = update(store, epoch, t, &data)
		lastData = &data
	}

	// try to get the last value
	epoch.Level = 0

	value, err = lookup.Lookup(lastData.Time, epoch, readFunc)
	if err != nil {
		t.Fatal(err)
	}

	if value != lastData {
		t.Fatalf("Expected lookup to return the last written value: %v. Got %v", lastData, value)
	}
}
