// ASH Extension of github.com/cbergoon/merkletree

package ash

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

type Content struct {
	//S string
	B []byte
}

//Equals tests for equality of two Contents
func (f Content) Equals(other Content) bool {
	if bytes.Compare(f.B, other.B) == 0 {
		return true
	} else {
		return false
	}
}

//MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
//a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root       *Node
	merkleRoot []byte
	Leafs      []*Node
}

//Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
//relationships, a hash, the content stored if it is a leaf, and other metadata.
type Node struct {
	Parent *Node
	Sister *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   []byte
	C      Content
}

//Computehash: Wrapper function for Keccak256
func Computehash(input []byte) (output []byte) {
	return crypto.Keccak256(bytes.TrimRight(input, "\x00"))
}

//verifyNode walks down the tree until hitting a leaf, calculating the hash at each level
//and returning the resulting hash of Node n.
func (n *Node) verifyNode() []byte {
	if n.leaf {
		return Computehash(n.C.B)
	}
	lhash := bytes.TrimRight(n.Left.verifyNode(), "\x00")
	rhash := bytes.TrimRight(n.Right.verifyNode(), "\x00")
	lr := append(lhash, rhash...)
	lrhash := Computehash(lr)
	fmt.Printf("L+R: %s => %x\n", bsplit(lr), lrhash)
	return lrhash

}

//calculateNodeHash is a helper function that calculates the hash of the node.
func (n *Node) calculateNodeHash() []byte {
	if n.leaf {
		return Computehash(n.C.B)
	}
	lrhash := n.calculateLRHash()
	return lrhash
}

// calculateLRHash is a helper function that calculates the hash given left and right node
func (n *Node) calculateLRHash() []byte {
	lhash := bytes.TrimRight(n.Left.Hash, "\x00")
	rhash := bytes.TrimRight(n.Right.Hash, "\x00")
	lr := append(lhash, rhash...)
	lrhash := Computehash(lr)
	//fmt.Printf("L+R: %s => %x\n", bsplit(lr), lrhash)
	return lrhash
}

//NewTree creates a new Merkle Tree using the content cs.
func NewTree(cs []Content) (*MerkleTree, error) {
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return nil, err
	}
	t := &MerkleTree{
		Root:       root,
		merkleRoot: root.Hash,
		Leafs:      leafs,
	}
	return t, nil
}

//buildWithContent is a helper function that for a given set of Contents, generates a
//corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
//Returns an error if cs contains no Contents.
func buildWithContent(cs []Content) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("Error: cannot construct tree with no content.")
	}
	var leafs []*Node
	for _, c := range cs {
		leafs = append(leafs, &Node{
			Hash: Computehash(c.B),
			C:    c,
			leaf: true,
		})
	}
	if len(leafs)%2 == 1 {
		leafs = append(leafs, leafs[len(leafs)-1])
		leafs[len(leafs)-1].dup = true
	}
	root := buildIntermediate(leafs)
	return root, leafs, nil
}

//buildIntermediate is a helper function that for a given list of leaf nodes, constructs
//the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate(nl []*Node) *Node {
	var nodes []*Node
	for i := 0; i < len(nl); i += 2 {
		chash := append(nl[i].Hash, nl[i+1].Hash...)
		h := Computehash(chash)
		n := &Node{
			Left:  nl[i],
			Right: nl[i+1],
			Hash:  h,
		}
		nodes = append(nodes, n)
		nl[i].Parent = n
		nl[i+1].Parent = n
		//TODO: set sister nodes
		n.Left.Sister = n.Right.Sister
		n.Right.Sister = n.Left.Sister
		if len(nl) == 2 {
			return n
		}
	}
	return buildIntermediate(nodes)
}

//MerkleRoot returns the unverified Merkle Root (hash of the root node) of the tree.
func (m *MerkleTree) MerkleRoot() []byte {
	return m.merkleRoot
}

//RebuildTree is a helper function that will rebuild the tree reusing only the content that
//it holds in the leaves.
func (m *MerkleTree) RebuildTree() error {
	var cs []Content
	for _, c := range m.Leafs {
		cs = append(cs, c.C)
	}
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

//RebuildTreeWith replaces the content of the tree and does a complete rebuild; while the root of
//the tree will be replaced the MerkleTree completely survives this operation. Returns an error if the
//list of content cs contains no entries.
func (m *MerkleTree) RebuildTreeWith(cs []Content) error {
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
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
	for j, l := range m.Leafs {
		if bytes.Compare(l.C.B, content) == 0 {
			currentSelf := l
			currentSister := l.Hash
			currentParent := l.Parent
			for currentParent != nil {
				if bytes.Compare(currentSelf.Hash, currentParent.Left.Hash) == 0 {
					currentSister = currentParent.Right.Hash
				} else {
					currentSister = currentParent.Left.Hash
				}
				mkproof = append(mkproof, currentSister...)
				//fmt.Printf("Self:%x | Sister:%x | Parent: %x\n", currentSelf.Hash, currentSister, currentParent.Hash)
				currentSelf = currentParent
				currentParent = currentParent.Parent
			}
			fmt.Printf("Content %s | Proof: %s\n", content, bsplit(mkproof))
			return true, m.merkleRoot, mkproof, int8(j)
		}
	}
	return false, nil, nil, index
}

func CheckProof(expectedMerkleRoot []byte, content []byte, mkproof []byte, index int8) (isValid bool, err error) {
	if len(mkproof)%32 != 0 {
		return false, errors.New("Invalid mkproof length")
	}

	/*
	   fmt.Printf("\n\n\nIndex is %v\n",index)
	   fmt.Printf("Proof is %s\n",bsplit(mkproof))
	*/
	totaldepth := len(mkproof) / 32
	merkleroot := append(content[:0], content...)
	merklepath := merkleroot

	depth := 0
	for depth < totaldepth {
		start := depth * 32
		end := start + 32

		if index%2 == 0 {
			c := make([]byte, 32)
			copy(c, mkproof[start:end])
			merkleroot = Computehash(append(merkleroot, c...))
			index = (index + 1) / 2
		} else {
			c := make([]byte, 32)
			copy(c, mkproof[start:end])
			merkleroot = Computehash(append(c, merkleroot...))
			index = index / 2
		}
		merklepath = append(merklepath, merkleroot...)
		depth++
	}

	if bytes.Compare(expectedMerkleRoot, merkleroot) != 0 {
		fmt.Printf("[CheckProof][FALSE] Expected: [%x] | Actual: [%x] | MRPath: {%v} | Proof {%v}\n", expectedMerkleRoot, merkleroot, bsplit(merklepath), bsplit(mkproof))
		return false, nil
	} else {
		fmt.Printf("[CheckProof][TRUE] MRPath: {%v} | Proof {%v}\n", bsplit(merklepath), bsplit(mkproof))
		return true, nil
	}
}


//Simple chunk split. Pad a chunk into 128 piece of 32 bytes
func chunksplit(chunk []byte) (segments [][]byte) {
	curr := 0
	for curr < 4096 {
		prev := curr
		curr += 32
		rawseg := make([]byte, 32)
		copy(rawseg[:], chunk[prev:curr])
		//rawseg := bytes.TrimRight(rawseg, "\x00")
		//fmt.Printf("Segemgt[%v:%v] | %v (%s)\n", prev, curr, rawseg, rawseg)
		segments = append(segments, rawseg)
	}
	return segments
}

//Compute segment index j given a secret
func getIndex(secret []byte) (index uint8) {
	seedhash := Computehash(secret)
	_ = binary.Read(bytes.NewReader(seedhash[31:]), binary.BigEndian, &index)
	index = index % 128
	fmt.Printf("SeedHash: %v | Index: %d\n", seedhash, index)
	return index
}

//Replace jth segment with h(content+seed)
func PrepareSegment(chunk []byte, secret []byte) (segments [][]byte) {
	j := getIndex(secret)
	segments = chunksplit(chunk)
	segments[j] = Computehash(append(segments[j], []byte(secret)...))
	return segments
}
