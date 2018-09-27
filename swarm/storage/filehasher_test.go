package storage

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/log"
)

const (
	segmentSize = 32
	branches    = 128
	chunkSize   = 4096
)

var pool *bmt.TreePool

var (
	start       = 14
	end         = 15
	dataLengths = []int{31, // 0
		32,                    // 1
		33,                    // 2
		63,                    // 3
		64,                    // 4
		65,                    // 5
		chunkSize,             // 6
		chunkSize + 31,        // 7
		chunkSize + 32,        // 8
		chunkSize + 63,        // 9
		chunkSize + 64,        // 10
		chunkSize * 2,         // 11
		chunkSize*2 + 32,      // 12
		chunkSize * 128,       // 13
		chunkSize*128 + 31,    // 14
		chunkSize*128 + 32,    // 15
		chunkSize*128 + 64,    // 16
		chunkSize * 129,       // 17
		chunkSize * 130,       // 18
		chunkSize * 128 * 128, // 19
	}
	expected = []string{
		"ece86edb20669cc60d142789d464d57bdf5e33cb789d443f608cbd81cfa5697d",
		"0be77f0bb7abc9cd0abed640ee29849a3072ccfd1020019fe03658c38f087e02",
		"3463b46d4f9d5bfcbf9a23224d635e51896c1daef7d225b86679db17c5fd868e",
		"95510c2ff18276ed94be2160aed4e69c9116573b6f69faaeed1b426fea6a3db8",
		"490072cc55b8ad381335ff882ac51303cc069cbcb8d8d3f7aa152d9c617829fe",
		"541552bae05e9a63a6cb561f69edf36ffe073e441667dbf7a0e9a3864bb744ea",
		"c10090961e7682a10890c334d759a28426647141213abda93b096b892824d2ef",
		"91699c83ed93a1f87e326a29ccd8cc775323f9e7260035a5f014c975c5f3cd28",
		"73759673a52c1f1707cbb61337645f4fcbd209cdc53d7e2cedaaa9f44df61285",
		"db1313a727ffc184ae52a70012fbbf7235f551b9f2d2da04bf476abe42a3cb42",
		"ade7af36ac0c7297dc1c11fd7b46981b629c6077bce75300f85b02a6153f161b",
		"29a5fb121ce96194ba8b7b823a1f9c6af87e1791f824940a53b5a7efe3f790d9",
		"61416726988f77b874435bdd89a419edc3861111884fd60e8adf54e2f299efd6",
		"3047d841077898c26bbe6be652a2ec590a5d9bd7cd45d290ea42511b48753c09",
		"e5c76afa931e33ac94bce2e754b1bb6407d07f738f67856783d93934ca8fc576",
		"485a526fc74c8a344c43a4545a5987d17af9ab401c0ef1ef63aefcc5c2c086df",
		"624b2abb7aefc0978f891b2a56b665513480e5dc195b4a66cd8def074a6d2e94",
		"b8e1804e37a064d28d161ab5f256cc482b1423d5cd0a6b30fde7b0f51ece9199",
		"59de730bf6c67a941f3b2ffa2f920acfaa1713695ad5deea12b4a121e5f23fa1",
		"522194562123473dcfd7a457b18ee7dee8b7db70ed3cfa2b73f348a992fdfd3b",
	}
)

func init() {
	pool = bmt.NewTreePool(sha3.NewKeccak256, 128, bmt.PoolSize*32)
}

func newAsyncHasher() bmt.SectionWriter {
	h := bmt.New(pool)
	return h.NewAsyncWriter(false)
}

func TestAltFileHasher(t *testing.T) {
	var mismatch int

	for i := start; i < end; i++ {
		dataLength := dataLengths[i]
		log.Info("start", "len", dataLength)
		fh := NewAltFileHasher(newAsyncHasher, 32, 128)
		_, data := generateSerialData(dataLength, 255, 0)
		l := 32
		offset := 0
		for i := 0; i < dataLength; i += 32 {
			remain := dataLength - offset
			if remain < l {
				l = remain
			}
			fh.Write(data[offset : offset+l])
			offset += 32
		}
		refHash := fh.Finish(nil)
		eq := true
		if expected[i] != fmt.Sprintf("%x", refHash) {
			mismatch++
			eq = false
		}
		t.Logf("[%7d+%4d]\t%v\tref: %x\texpect: %s", dataLength/chunkSize, dataLength%chunkSize, eq, refHash, expected[i])
	}
	if mismatch > 0 {
		t.Fatalf("mismatches: %d/%d", mismatch, len(dataLengths))
	}
}

func TestReferenceFileHasher(t *testing.T) {
	h := bmt.New(pool)
	var mismatch int
	for i := start; i < end; i++ {
		dataLength := dataLengths[i]
		log.Info("start", "len", dataLength)
		fh := NewReferenceFileHasher(h, 128)
		_, data := generateSerialData(dataLength, 255, 0)
		refHash := fh.Hash(bytes.NewReader(data), len(data)).Bytes()
		eq := true
		if expected[i] != fmt.Sprintf("%x", refHash) {
			mismatch++
			eq = false
		}
		t.Logf("[%7d+%4d]\t%v\tref: %x\texpect: %s", dataLength/chunkSize, dataLength%chunkSize, eq, refHash, expected[i])
	}
	if mismatch > 0 {
		t.Fatalf("mismatches: %d/%d", mismatch, len(dataLengths))
	}
}

func TestSum(t *testing.T) {

	var mismatch int
	serialOffset := 0

	for i := start; i < end; i++ {
		dl := dataLengths[i]
		chunks := dl / chunkSize
		log.Debug("testing", "c", chunks, "s", dl%chunkSize)
		fhStartTime := time.Now()
		fh := NewFileHasher(newAsyncHasher, 128, 32)
		_, data := generateSerialData(dl, 255, serialOffset)
		for i := 0; i < len(data); i += 32 {
			max := i + 32
			if len(data) < max {
				max = len(data)
			}
			_, err := fh.WriteBuffer(i, data[i:max])
			if err != nil {
				t.Fatal(err)
			}
		}

		fh.SetLength(int64(dl))
		h := fh.Sum(nil)
		rhStartTime := time.Now()
		rh := NewReferenceFileHasher(bmt.New(pool), 128)
		p := rh.Hash(bytes.NewReader(data), len(data)).Bytes()
		rhDur := time.Now().Sub(rhStartTime)

		eq := bytes.Equal(p, h)
		if !eq {
			mismatch++
		}
		t.Logf("[%3d + %2d]\t%v\t%x\t%x", chunks, dl%chunkSize, eq, p, h)
		t.Logf("ptime %v\tftime %v", rhDur, rhStartTime.Sub(fhStartTime))
	}
	if mismatch > 0 {
		t.Fatalf("%d/%d mismatches", mismatch, len(dataLengths))
	}
}
