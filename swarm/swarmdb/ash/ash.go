package ash

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

type Content struct {
	B []byte
}

type MerkleTree struct {
	Roothash []byte
	Root     *Node
	Leafs    []*Node
}

type Node struct {
	Parent      *Node
	Sister      *Node
	Left        *Node
	Right       *Node
	Hash        []byte
	Cont        Content
	isLeaf      bool
	isDuplicate bool
}

type AshChallenge struct {
	ProofRequired bool `json: "proofrequired"`
	Index         int8 `json: index`
}

type AshRequest struct {
	ChunkID   []byte `json:"chunkID"`
	Seed      []byte `json: seed`
	Challenge *AshChallenge
}

type AshResponse struct {
	root  []byte `json: "-"`
	Mash  string `json: "mash"`
	Proof *MerkleProof
}

type MerkleProof struct {
	Root  []byte `json: "root"`
	Path  []byte `json: "path"`
	Index int8   `json: "index"`
}

func (u *MerkleProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		&struct {
			Root  string `json: "root"`
			Path  string `json: "path"`
			Index int8   `json: "index"`
		}{
			Root:  hex.EncodeToString(u.Root),
			Path:  hex.EncodeToString(u.Path),
			Index: u.Index,
		})
}

func NewTree(contents []Content) (m *MerkleTree, err error) {
	root, leafs, err := buildWithContent(contents)
	if err != nil {
		return nil, err
	}
	m = &MerkleTree{
		Root:     root,
		Roothash: root.Hash,
		Leafs:    leafs,
	}
	return m, nil
}

func buildWithContent(contents []Content) (*Node, []*Node, error) {
	if len(contents) == 0 {
		return nil, nil, errors.New("Error: cannot construct tree with no content.")
	}
	var leafs []*Node
	for _, content := range contents {
		leafs = append(leafs, &Node{
			Cont:   content,
			Hash:   Computehash(content.B),
			isLeaf: true,
		})
	}
	if len(leafs)%2 == 1 {
		leafs = append(leafs, leafs[len(leafs)-1])
		leafs[len(leafs)-1].isDuplicate = true
	}
	//TODO: n = 2
	root := buildIntermediate(leafs)
	return root, leafs, nil
}

func buildIntermediate(leafs []*Node) *Node {
	var nodes []*Node
	for i := 0; i < len(leafs); i += 2 {
		parentN := &Node{
			Left:  leafs[i],
			Right: leafs[i+1],
			Hash:  Computehash(append(leafs[i].Hash, leafs[i+1].Hash...)),
		}
		nodes = append(nodes, parentN)
		leafs[i].Parent = parentN
		leafs[i+1].Parent = parentN
		//TODO: set sister nodes
		parentN.Left.Sister = parentN.Right.Sister
		parentN.Right.Sister = parentN.Left.Sister
		if len(leafs) == 2 {
			return parentN
		}
	}
	return buildIntermediate(nodes)
}

func (m *MerkleTree) RebuildTree() error {
	var contents []Content
	for _, leaf := range m.Leafs {
		contents = append(contents, leaf.Cont)
	}
	root, leafs, err := buildWithContent(contents)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.Roothash = root.Hash
	return nil
}

func (m *MerkleTree) RebuildTreeWith(contents []Content) error {
	root, leafs, err := buildWithContent(contents)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.Roothash = root.Hash
	return nil
}

func bsplit(rawhash []byte) string {
	var segments [][]byte
	n := 0
	for n < len(rawhash)/32 {
		segments = append(segments, rawhash[n*32:(n+1)*32])
		n++
	}
	s := fmt.Sprintf("%x", segments)
	return s
}

func (m *MerkleTree) GetProof(content []byte) (ok bool, mkroot []byte, proof []byte, index int8) {
	var mkproof []byte
	for j, leaf := range m.Leafs {
		if bytes.Compare(leaf.Cont.B, content) == 0 {
			currentSelf := leaf
			currentSister := leaf.Hash
			currentParent := leaf.Parent
			mkproof = append(mkproof, content...)
			for currentParent != nil {
				if bytes.Compare(currentSelf.Hash, currentParent.Left.Hash) == 0 {
					currentSister = currentParent.Right.Hash
				} else {
					currentSister = currentParent.Left.Hash
				}
				mkproof = append(mkproof, currentSister...)
				currentSelf = currentParent
				currentParent = currentParent.Parent
			}
			return true, m.Roothash, mkproof, int8(j)
		}
	}
	return false, nil, nil, index
}

func CheckProof(expectedMerkleRoot []byte, mkproof []byte, index int8) (isValid bool, merkleroot []byte, err error) {
	if len(mkproof)%32 != 0 {
		return false, merkleroot, errors.New("Invalid mkproof length")
	}
	content := make([]byte, 32)
	copy(content, mkproof[0:36])
	merkleroot = Computehash(append(content[:0], content...))
	merklepath := merkleroot
	depth, totaldepth := 1, len(mkproof)/32
	for depth < totaldepth {
		start := depth * 32
		end := start + 32

		if index%2 == 0 {
			rhash := make([]byte, 32)
			copy(rhash, mkproof[start:end])
			merkleroot = Computehash(append(merkleroot, rhash...))
			index = (index + 1) / 2
		} else {
			rhash := make([]byte, 32)
			copy(rhash, mkproof[start:end])
			merkleroot = Computehash(append(rhash, merkleroot...))
			index = index / 2
		}
		merklepath = append(merklepath, merkleroot...)
		depth++
	}
	if bytes.Compare(expectedMerkleRoot, merkleroot) != 0 {
		return false, merkleroot, nil
	} else {
		return true, merkleroot, nil
	}
}

//Computehash: Wrapper function for Hash function
func Computehash(input []byte) (output []byte) {
	return crypto.Keccak256(bytes.TrimRight(input, "\x00"))
}

//Simple chunk split. Pad a chunk into 128 piece of 32 bytes
func chunksplit(chunk []byte) (rawseg []Content) {
	curr := 0
	for curr < 4096 {
		prev := curr
		curr += 32
		seg := make([]byte, 32)
		copy(seg[:], chunk[prev:curr])
		rawseg = append(rawseg, Content{B: seg})
	}
	return rawseg
}

//Compute segment index j given a secret
func getIndex(secret []byte) (index uint8) {
	seedhash := Computehash(secret)
	_ = binary.Read(bytes.NewReader(seedhash[31:]), binary.BigEndian, &index)
	index = index % 128
	return index
}

//Replace jth segment with h(content+seed)
func PrepareSegment(chunk []byte, secret []byte) (segments []Content) {
	j := getIndex(secret)
	rawseg := chunksplit(chunk)
	rawseg[j].B = Computehash(append(rawseg[j].B, []byte(secret)...))
	return rawseg
}

//GenerateAsh returns a merkle root based on given seed
func GenerateAsh(seed []byte, rawchunk []byte) (rootHash []byte, err error) {
    segments := PrepareSegment(rawchunk, seed)
    tree, err := NewTree(segments)
    if err != nil {
        return rootHash, err
    }
    return tree.Roothash, nil
}

//ComputeAsh answers to an AshRequest
func ComputeAsh(request AshRequest, rawchunk []byte) (response AshResponse, err error) {
	segments := PrepareSegment(rawchunk, request.Seed)
	tree, err := NewTree(segments)
	if err != nil {
		return response, err
	}
	challenge := request.Challenge
	if challenge.ProofRequired {
		ok, merkleroot, mkproof, jth := tree.GetProof(segments[challenge.Index].B)
		if !ok {
			return response, err
		}
		response.Proof = &MerkleProof{Root: merkleroot, Path: mkproof, Index: jth}
	}
	response.root = tree.Roothash
	response.Mash = fmt.Sprintf("%x", response.root)
	return response, nil
}

//Verify AshResponse
func VerifyAsh(response AshResponse, expectedMash []byte) (isValid bool, givenRoot []byte, err error) {
	if response.Proof != nil {
		merkleproof := response.Proof
		ok, merkleroot, err := CheckProof(merkleproof.Root, merkleproof.Path, merkleproof.Index)
        if err != nil {
            return false, givenRoot, err
        }
		if !ok || bytes.Compare(response.root, merkleroot) != 0 {
			return false, merkleroot, nil
		}
	}
	mash := Computehash(response.root)
	if bytes.Compare(mash, expectedMash) != 0 {
		return false, response.root, nil
	}
	return true, response.root, nil
}
