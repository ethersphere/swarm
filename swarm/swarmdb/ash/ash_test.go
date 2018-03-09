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
	isValidProof bool
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
	}
	return segments
}

func (data ByteData) CreateContent() (segments []Content) {
	modifiedsegments := PrepareSegment(data.b, []byte(data.seed))
	return modifiedsegments
}

func LoadTestChunk(path string) []byte {
	data := make([]byte, 4096)
	d, _ := ioutil.ReadFile(path)
	copy(data[:], d)
	return data
}

var testset = []Testdata{
	{
		//Merkle proof of c
		data: StringData{
			s: []string{"a", "b", "c", "d"},
		},
		isValidProof: true,
		index:        3,
		expectedHash: []byte{104, 32, 63, 144, 233, 208, 125, 197, 133, 146, 89, 215, 83, 110, 135, 166, 186, 157, 52, 95, 37, 82, 181, 185, 222, 41, 153, 221, 206, 156, 225, 191},
		merkleproof:  []byte{100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 66, 182, 57, 60, 31, 83, 6, 15, 227, 221, 191, 205, 122, 173, 204, 168, 148, 70, 90, 90, 67, 143, 105, 200, 125, 121, 11, 34, 153, 185, 178, 128, 91, 33, 216, 70, 177, 137, 239, 174, 176, 55, 125, 107, 176, 210, 1, 179, 135, 42, 54, 62, 96, 124, 37, 8, 143, 2, 91, 12, 106, 225, 248},
	},
	{
		//Merkle proof of f
		data: StringData{
			s: []string{"a", "b", "c", "d", "e", "f", "g", "h"},
		},
		index:        5,
		isValidProof: true,
		expectedHash: []byte{205, 7, 39, 47, 73, 85, 221, 207, 218, 195, 143, 243, 109, 255, 157, 62, 67, 83, 73, 137, 35, 103, 154, 181, 72, 186, 135, 227, 70, 72, 228, 163},
		merkleproof:  []byte{102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 168, 152, 44, 137, 216, 9, 135, 251, 154, 81, 14, 37, 152, 30, 233, 23, 2, 6, 190, 33, 175, 60, 142, 14, 179, 18, 239, 29, 51, 130, 231, 97, 225, 138, 92, 46, 229, 32, 46, 205, 239, 237, 104, 63, 3, 20, 91, 19, 67, 48, 77, 190, 208, 26, 236, 185, 64, 50, 183, 248, 1, 132, 79, 10, 104, 32, 63, 144, 233, 208, 125, 197, 133, 146, 89, 215, 83, 110, 135, 166, 186, 157, 52, 95, 37, 82, 181, 185, 222, 41, 153, 221, 206, 156, 225, 191},
	},
	{
		//Merkle proof of segment 127, with salt "Hello"
		data: ByteData{
			b:    LoadTestChunk("data1"),
			seed: "Hello",
		},
		isValidProof: true,
		index:        127,
		expectedHash: []byte{173, 74, 248, 176, 39, 141, 144, 117, 40, 135, 152, 232, 228, 147, 196, 76, 175, 165, 222, 31, 89, 233, 54, 2, 210, 251, 231, 123, 238, 249, 23, 15},
		merkleproof:  []byte{32, 101, 116, 101, 114, 110, 97, 108, 32, 101, 116, 104, 105, 99, 115, 32, 101, 118, 105, 100, 101, 110, 99, 101, 32, 101, 118, 105, 108, 32, 101, 118, 200, 16, 210, 30, 42, 215, 164, 110, 135, 131, 139, 141, 132, 46, 144, 97, 244, 163, 16, 159, 23, 246, 217, 67, 80, 46, 226, 53, 163, 161, 33, 91, 207, 116, 2, 168, 129, 254, 164, 132, 87, 136, 197, 232, 118, 0, 53, 71, 252, 7, 141, 241, 120, 176, 226, 52, 33, 66, 184, 158, 159, 215, 175, 229, 38, 38, 123, 166, 116, 119, 26, 121, 21, 151, 181, 170, 103, 228, 17, 17, 225, 147, 132, 31, 87, 248, 131, 254, 57, 93, 32, 223, 13, 134, 2, 157, 127, 39, 201, 221, 208, 171, 29, 61, 78, 137, 71, 163, 71, 140, 118, 70, 122, 136, 15, 244, 141, 26, 131, 233, 51, 97, 189, 225, 157, 153, 150, 217, 11, 110, 242, 36, 92, 165, 79, 220, 122, 195, 129, 221, 58, 2, 3, 213, 120, 225, 35, 36, 91, 42, 31, 168, 39, 22, 24, 104, 229, 29, 108, 74, 66, 89, 190, 38, 111, 90, 248, 189, 228, 8, 115, 102, 145, 24, 111, 76, 36, 85, 41, 5, 116, 142, 226, 125, 208, 30, 87, 232, 219, 126, 123, 34, 98, 198, 33, 61, 180, 158, 130, 54, 200, 251, 243, 136, 163, 12, 10, 228, 226, 123, 217, 167, 141, 79, 79, 17, 87, 207, 223, 113, 250, 214, 77, 250},
	},
	{
		//Merkle proof of an empty segment, with salt "World"
		data: ByteData{
			b:    LoadTestChunk("data2"),
			seed: "World",
		},
		isValidProof: true,
		index:        127,
		expectedHash: []byte{48, 164, 230, 206, 37, 152, 183, 21, 80, 23, 109, 246, 253, 73, 83, 124, 7, 166, 168, 250, 61, 212, 40, 6, 155, 220, 65, 180, 52, 140, 103, 107},
		merkleproof:  []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 109, 241, 249, 143, 103, 177, 69, 185, 117, 129, 5, 156, 143, 229, 127, 35, 104, 80, 250, 102, 12, 132, 104, 202, 224, 238, 143, 0, 14, 106, 49, 245, 207, 116, 2, 168, 129, 254, 164, 132, 87, 136, 197, 232, 118, 0, 53, 71, 252, 7, 141, 241, 120, 176, 226, 52, 33, 66, 184, 158, 159, 215, 175, 229, 38, 38, 123, 166, 116, 119, 26, 121, 21, 151, 181, 170, 103, 228, 17, 17, 225, 147, 132, 31, 87, 248, 131, 254, 57, 93, 32, 223, 13, 134, 2, 157, 127, 39, 201, 221, 208, 171, 29, 61, 78, 137, 71, 163, 71, 140, 118, 70, 122, 136, 15, 244, 141, 26, 131, 233, 51, 97, 189, 225, 157, 153, 150, 217, 11, 110, 242, 36, 92, 165, 79, 220, 122, 195, 129, 221, 58, 2, 3, 213, 120, 225, 35, 36, 91, 42, 31, 168, 39, 22, 24, 104, 229, 29, 108, 74, 111, 71, 160, 240, 245, 36, 79, 11, 233, 207, 212, 170, 0, 17, 78, 127, 210, 101, 91, 89, 73, 66, 161, 0, 7, 82, 52, 122, 234, 5, 74, 96, 150, 119, 34, 133, 246, 14, 149, 171, 34, 133, 10, 191, 170, 16, 79, 207, 148, 134, 169, 133, 139, 108, 102, 31, 175, 106, 63, 226, 7, 91, 151, 92},
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
		mismatch := bytes.Compare(tree.Roothash, testset[i].expectedHash) != 0
		if mismatch && testset[i].isValidProof {
			t.Errorf("ERROR: Expected MerkleRoot [%v] and Actual MerkleRoot [%v] to be the same", testset[i].expectedHash, tree.Roothash)
		} else if !mismatch && !testset[i].isValidProof {
			t.Errorf("ERROR: Expect mismatch between MerkleRoot [%v] and Actual MerkleRoot [%v]", testset[i].expectedHash, tree.Roothash)
		} else {
			fmt.Printf("Success: MerkleRoot [%x]\n", tree.Roothash)
		}
	}
}

func TestCheckProof(t *testing.T) {
	for i := 0; i < len(testset); i++ {
		testset[i].segments = testset[i].data.CreateContent()
		isValid, mroot, err := CheckProof(testset[i].expectedHash, testset[i].merkleproof, testset[i].index)
		if err != nil {
			if isValid {
				t.Error("Uncaught ERROR:  ", err)
			} else {
				t.Error("ERROR:  ", err)
			}
		}
		if isValid == testset[i].isValidProof {
			fmt.Printf("Success: checkProof Verified [%x]\n", testset[i].expectedHash)
		} else {
			t.Error("ERROR: CheckProof Exptected [%x]| Actual [%x]\n", testset[i].expectedHash, mroot)
		}
	}
}

func TestGetProof(t *testing.T) {
	for i := 0; i < len(testset); i++ {
		testset[i].segments = testset[i].data.CreateContent()
		tree, err := NewTree(testset[i].segments)
		if err != nil {
			t.Error("Uncaught ERROR:  ", err)
		}
		for j, segment := range testset[i].segments {
			ok, merkleroot, mkproof, index := tree.GetProof(segment.B)
			if ok != testset[i].isValidProof {
				t.Error("ERROR: Invalid MerkleRoot [%x] | Proof [%x] | Index [%d]\n", testset[i].merkleproof, index)
			} else if int8(j) == testset[i].index {
				fmt.Printf("Success: MerkleRoot [%x] | Proof [%s] | Index [%d]\n", merkleroot, bsplit(mkproof), index)
			}
		}
	}
}

func TestGetProofAndCheckProof(t *testing.T) {
	for i := 0; i < len(testset); i++ {
		var testproof []TestProof
		testset[i].segments = testset[i].data.CreateContent()
		tree, err := NewTree(testset[i].segments)
		if err != nil {
			t.Error("Uncaught ERROR:  ", err)
		}
		for _, segment := range testset[i].segments {
			content := segment.B
			chash := Computehash(content)
			ok, merkleroot, mkproof, ind := tree.GetProof(content)
			if !ok {
				t.Error("ERROR: Invalid MerkleRoot [%x] | Proof [%x]\n", merkleroot, testset[i].merkleproof)
			} else {
				testproof = append(testproof, TestProof{expectedHash: merkleroot, contentHash: chash, merkleproof: mkproof, index: ind})
				fmt.Printf("Success: MerkleRoot [%x] | Content [%x]| Proof [%x]\n", merkleroot, chash, mkproof)
			}
		}
		for k, proof := range testproof {
			isValid, mroot, err := CheckProof(proof.expectedHash, proof.merkleproof, proof.index)
			if err != nil {
				if isValid {
					t.Error("Uncaught ERROR:  ", err)
				} else {
					t.Error("ERROR:  ", err)
				}
			}
			if !isValid {
				t.Error("ERROR: CheckProof Exptected [%x]| Actual [%x]", proof.expectedHash, mroot)
			} else if k%32 == 0 {
				fmt.Printf("Success: checkProof Verified [%x]\n", mroot)
			}

		}
	}
}
