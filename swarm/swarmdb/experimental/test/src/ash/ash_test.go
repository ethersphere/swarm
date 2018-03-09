package ash

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

type Testdata struct {
	data         Data
	segments     []Content
	seed         string
	expectedHash []byte
	index        int8
	merkleproof  []byte
}

type TestProof struct {
	expectedHash []byte
	contentHash  []byte
	merkleproof  []byte
	index        int8
}

type Data interface {
	CreateContent() []Content
}

type StringData struct {
	s []string
}

type ByteData struct {
	b    []byte
	seed string
}

func (data StringData) CreateContent() (segments []Content) {
	for _, rs := range data.s {
		rawseg := make([]byte, 32)
		copy(rawseg[:], rs)
		segments = append(segments, Content{B: rawseg})
		//segments = append(segments, Content{S: rs, B: rawseg})
	}
	return segments
}

func (data ByteData) CreateContent() (segments []Content) {
	modifiedsegments := PrepareSegment(data.b, []byte(data.seed))
	for _, seg := range modifiedsegments {
		segments = append(segments, Content{B: seg})
		//segments = append(segments, Content{S: fmt.Sprintf("%s", seg), B: seg})
	}
	return segments
}

func LoadTestChunk(path string) []byte {
	data := make([]byte, 4096)
	d, _ := ioutil.ReadFile(path)
	copy(data[:], d)
	return data
}

var testset = []Testdata{
	{
		data: StringData{
			s: []string{"a", "b", "c", "d"},
		},
		expectedHash: []byte{104, 32, 63, 144, 233, 208, 125, 197, 133, 146, 89, 215, 83, 110, 135, 166, 186, 157, 52, 95, 37, 82, 181, 185, 222, 41, 153, 221, 206, 156, 225, 191},
		index:        0,
		merkleproof:  []byte{181, 85, 61, 227, 21, 224, 237, 245, 4, 217, 21, 10, 248, 45, 175, 165, 196, 102, 127, 166, 24, 237, 10, 111, 25, 198, 155, 65, 22, 108, 85, 16, 210, 83, 165, 45, 76, 176, 13, 226, 137, 94, 133, 242, 82, 158, 41, 118, 230, 170, 170, 92, 24, 16, 107, 104, 171, 102, 129, 62, 20, 65, 86, 105},
	},
	{
		data: StringData{
			s: []string{"a", "b", "c", "d", "e", "f", "g", "h"},
		},
		expectedHash: []byte{205, 7, 39, 47, 73, 85, 221, 207, 218, 195, 143, 243, 109, 255, 157, 62, 67, 83, 73, 137, 35, 103, 154, 181, 72, 186, 135, 227, 70, 72, 228, 163},
		index:        0,
		merkleproof:  []byte{181, 85, 61, 227, 21, 224, 237, 245, 4, 217, 21, 10, 248, 45, 175, 165, 196, 102, 127, 166, 24, 237, 10, 111, 25, 198, 155, 65, 22, 108, 85, 16, 210, 83, 165, 45, 76, 176, 13, 226, 137, 94, 133, 242, 82, 158, 41, 118, 230, 170, 170, 92, 24, 16, 107, 104, 171, 102, 129, 62, 20, 65, 86, 105, 243, 19, 252, 158, 177, 196, 134, 75, 27, 142, 120, 41, 102, 86, 251, 120, 49, 204, 14, 212, 99, 97, 191, 52, 82, 219, 28, 76, 236, 67, 0, 80},
	},
	{
		data: ByteData{
			b:    LoadTestChunk("data1"),
			seed: "Hello",
		},
		expectedHash: []byte{173, 74, 248, 176, 39, 141, 144, 117, 40, 135, 152, 232, 228, 147, 196, 76, 175, 165, 222, 31, 89, 233, 54, 2, 210, 251, 231, 123, 238, 249, 23, 15},
		index:        0,
		merkleproof:  []byte{165, 119, 242, 218, 32, 143, 84, 28, 244, 35, 208, 217, 149, 168, 239, 151, 35, 96, 250, 83, 156, 41, 29, 223, 125, 21, 9, 65, 233, 189, 226, 41, 61, 20, 65, 245, 40, 42, 84, 211, 68, 63, 130, 28, 250, 40, 193, 41, 219, 136, 190, 187, 1, 240, 60, 59, 48, 10, 182, 146, 199, 80, 84, 198, 33, 107, 159, 107, 235, 236, 215, 104, 178, 24, 200, 55, 144, 189, 70, 206, 159, 61, 82, 162, 26, 89, 169, 125, 35, 7, 242, 210, 246, 203, 230, 51, 24, 85, 97, 121, 52, 28, 241, 249, 224, 65, 250, 101, 121, 88, 22, 65, 253, 146, 13, 95, 98, 22, 153, 182, 214, 231, 3, 75, 6, 167, 218, 117, 61, 25, 93, 9, 244, 72, 250, 14, 39, 116, 79, 140, 86, 76, 185, 87, 244, 237, 30, 56, 218, 8, 27, 166, 232, 179, 4, 236, 120, 117, 46, 180, 166, 214, 182, 91, 237, 228, 38, 150, 111, 106, 220, 118, 8, 43, 197, 158, 76, 14, 42, 127, 74, 89, 21, 65, 119, 105, 73, 170, 15, 28, 20, 175, 39, 36, 79, 80, 67, 131, 109, 191, 193, 58, 244, 213, 6, 86, 49, 40, 116, 29, 196, 27, 80, 95, 251, 224, 74, 99, 76, 243, 162, 39, 112, 144},
	},
	{
		data: ByteData{
			b:    LoadTestChunk("data2"),
			seed: "World",
		},
		expectedHash: []byte{48, 164, 230, 206, 37, 152, 183, 21, 80, 23, 109, 246, 253, 73, 83, 124, 7, 166, 168, 250, 61, 212, 40, 6, 155, 220, 65, 180, 52, 140, 103, 107},
		index:        0,
		merkleproof:  []byte{165, 119, 242, 218, 32, 143, 84, 28, 244, 35, 208, 217, 149, 168, 239, 151, 35, 96, 250, 83, 156, 41, 29, 223, 125, 21, 9, 65, 233, 189, 226, 41, 61, 20, 65, 245, 40, 42, 84, 211, 68, 63, 130, 28, 250, 40, 193, 41, 219, 136, 190, 187, 1, 240, 60, 59, 48, 10, 182, 146, 199, 80, 84, 198, 33, 107, 159, 107, 235, 236, 215, 104, 178, 24, 200, 55, 144, 189, 70, 206, 159, 61, 82, 162, 26, 89, 169, 125, 35, 7, 242, 210, 246, 203, 230, 51, 24, 85, 97, 121, 52, 28, 241, 249, 224, 65, 250, 101, 121, 88, 22, 65, 253, 146, 13, 95, 98, 22, 153, 182, 214, 231, 3, 75, 6, 167, 218, 117, 42, 96, 113, 222, 14, 194, 64, 150, 173, 166, 98, 222, 182, 107, 230, 156, 218, 116, 35, 213, 185, 51, 144, 79, 34, 208, 225, 73, 158, 253, 53, 197, 166, 214, 182, 91, 237, 228, 38, 150, 111, 106, 220, 118, 8, 43, 197, 158, 76, 14, 42, 127, 74, 89, 21, 65, 119, 105, 73, 170, 15, 28, 20, 175, 184, 51, 157, 79, 115, 100, 248, 104, 72, 232, 228, 118, 115, 67, 25, 126, 49, 100, 241, 134, 76, 107, 127, 205, 66, 70, 76, 86, 34, 216, 193, 98},
	},
}

func TestNewTree(t *testing.T) {
	for i := 0; i < len(testset); i++ {
		testset[i].segments = testset[i].data.CreateContent()
		fmt.Printf("Segment: %s\n", testset[i].segments)
		tree, err := NewTree(testset[i].segments)
		if err != nil {
			t.Error("Uncaught ERROR:  ", err)
		}
		if bytes.Compare(tree.MerkleRoot(), testset[i].expectedHash) != 0 {
			t.Errorf("ERROR: Expected MerkleRoot [%v], Actual MerkleRoot [%v]", testset[i].expectedHash, tree.MerkleRoot())
		} else {
			fmt.Printf("Success: MerkleRoot [%x]\n", tree.MerkleRoot())
		}
	}
}

func TestCheckProof(t *testing.T) {
	for i := 0; i < len(testset); i++ {
		testset[i].segments = testset[i].data.CreateContent()
		contentHash := Computehash(testset[i].segments[0].B)
		isValid, err := CheckProof(testset[i].expectedHash, contentHash, testset[i].merkleproof, testset[i].index)
		if err != nil {
			if isValid {
				t.Error("Uncaught ERROR:  ", err)
			} else {
				t.Error("ERROR:  ", err)
			}
		}
		if isValid {
			fmt.Printf("Success: checkProof Verified [%x]\n", testset[i].expectedHash)
		} else {
			t.Error("ERROR: CheckProof Exptected [%x]| Actual [%x]\n", testset[i].expectedHash, contentHash)
		}
	}
}

func TestGetProof(t *testing.T) {
	for i := 0; i < 1; i++ {
		testset[i].segments = testset[i].data.CreateContent()
		tree, err := NewTree(testset[i].segments)
		if err != nil {
			t.Error("Uncaught ERROR:  ", err)
		}
		for j, segment := range testset[i].segments {
			ok, merkleroot, mkproof, index := tree.GetProof(segment.B)
			if !ok {
				t.Error("ERROR: Invalid MerkleRoot [%x] | Proof [%x] | Index [%d]\n", testset[i].merkleproof, index)
			} else if j%128 == 0 {
				fmt.Printf("Success: MerkleRoot [%x] | Proof [%s] | Index [%d]\n", merkleroot, bsplit(mkproof), index)
			}
		}
	}
}

func TestGetProofAndCheckProof(t *testing.T) {
	for i := 0; i < 4; i++ {
		var testproof []TestProof
		testset[i].segments = testset[i].data.CreateContent()
		tree, err := NewTree(testset[i].segments)
		if err != nil {
			t.Error("Uncaught ERROR:  ", err)
		}
		expectedroot := testset[i].expectedHash
		for _, segment := range testset[i].segments {
			content := segment.B
			chash := Computehash(content)
			ok, merkleroot, mkproof, ind := tree.GetProof(content)
			if !ok {
				t.Error("ERROR: Invalid MerkleRoot [%x] | Proof [%x]\n", merkleroot, testset[i].merkleproof)
			} else {
				testproof = append(testproof, TestProof{expectedHash: merkleroot, contentHash: chash, merkleproof: mkproof, index: ind})
				fmt.Printf("Success: MerkleRoot [%x] | Content [%x]| Proof [%v]\n", merkleroot, chash, mkproof)
			}
		}

		for k, proof := range testproof {
			isValid, err := CheckProof(proof.expectedHash, proof.contentHash, proof.merkleproof, proof.index)
			if err != nil {
				if isValid {
					t.Error("Uncaught ERROR:  ", err)
				} else {
					t.Error("ERROR:  ", err)
				}
			}
			if !isValid {
				t.Error("ERROR: CheckProof Exptected [%x]| Actual [%x]", expectedroot, proof.expectedHash)
			} else if k%32 == 0 {
				fmt.Printf("Success: checkProof Verified\n")
			}

		}
	}
}
