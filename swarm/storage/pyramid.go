// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

const (
	processors = 1
)

type Tree struct {
	Chunks int64
	Depth  int64
	Levels []map[int64]*Node
	Lock   sync.RWMutex
}

type Node struct {
	Pending  int64
	Size     uint64
	Children []common.Hash
	Last     bool
}

func (self *Node) String() string {
	var children []string
	for _, node := range self.Children {
		children = append(children, node.Hex())
	}
	return fmt.Sprintf("pending: %v, size: %v, last :%v, children: %v", self.Pending, self.Size, self.Last, strings.Join(children, ", "))
}

type Task struct {
	Index int64 // Index of the chunk being processed
	Size  uint64
	Data  []byte // Binary blob of the chunk
	Last  bool
}

type PyramidChunker struct {
	hashFunc    Hasher
	chunkSize   int64
	hashSize    int64
	branches    int64
	workerCount int
}

func NewPyramidChunker(params *ChunkerParams) (self *PyramidChunker) {
	self = &PyramidChunker{}
	self.hashFunc = MakeHashFunc(params.Hash)
	self.branches = params.Branches
	self.hashSize = int64(self.hashFunc().Size())
	self.chunkSize = self.hashSize * self.branches
	self.workerCount = 1
	return
}

//TODO: this is a fake Join so it compiles
func (self *PyramidChunker) Join(key Key, chunkC chan *Chunk) LazySectionReader {

	return &LazyChunkReader{
		key:       key,
		chunkC:    chunkC,
		chunkSize: self.chunkSize,
		branches:  self.branches,
		hashSize:  self.hashSize,
	}
}

func (self *PyramidChunker) Split(data io.Reader, size int64, chunkC chan *Chunk, swg, wwg *sync.WaitGroup) (Key, error) {

	//prepare results tree
	results := Tree{
		Depth:  0,
		Levels: make([]map[int64]*Node, 1),
	}

  //initialize the results, allocate lowest level or pyramid
	results.Levels[0] = make(map[int64]*Node)

	// Create a pool of workers to crunch through the file
	tasks := make(chan *Task, 2*processors)
	pend := new(sync.WaitGroup)
	abortC := make(chan bool)
	for i := 0; i < processors; i++ {
		pend.Add(1)
		go self.processor(pend, swg, tasks, chunkC, &results)
	}

	// Feed the chunks into the task pool
	read := 0
	for index := 0; ; index++ {
		//create the byte array buffer where chunks will be read into; reserve 8 bytes (64bits) for data length
		buffer := make([]byte, self.chunkSize+8)
		//read n-amount of bytes into buffer, starting after 8 bytes
		n, err := data.Read(buffer[8:])
		//increase total amount of read bytes
		read += n
		//evaluate if we completed the last chunk
		last := int64(read) == size || err == io.ErrUnexpectedEOF || err == io.EOF
		if err != nil && !last {
			//something went wrong...
			close(abortC)
			break
		}
		//write length or data into first 8 bytes
		binary.LittleEndian.PutUint64(buffer[:8], uint64(n))
		pend.Add(1)
		select {
		case tasks <- &Task{Index: int64(index), Size: uint64(n), Data: buffer[:n+8], Last: last}:
		case <-abortC:
			return nil, err
		}
		if last {
			break
		}
	}
	// Wait for the workers and return
	close(tasks)
	pend.Wait()

	//The top of the pyramid is at results.Depth
	key := results.Levels[results.Depth][0].Children[0][:]
	return key, nil
}

func (self *PyramidChunker) processor(pend, swg *sync.WaitGroup, tasks chan *Task, chunkC chan *Chunk, results *Tree) {
	defer pend.Done()

	// Start processing leaf chunks ad infinitum
	hasher := self.hashFunc()

	for task := range tasks {
		size := task.Size
		data := task.Data
		var node *Node

		// New chunk received, reset the hasher and start processing
    // Hash leaf nodes as they come in
		hasher.Reset()
		hasher.Write(task.Data)
		hash := hasher.Sum(nil)

		last := task.Last

    //create nodes one by one; we don't know how many we will get...
		node = &Node{0, 0, make([]common.Hash, 1), last}
		d := 0
    //set the node's Children as the data's hash values
		copy(node.Children[d][:], hash)
		node.Size += size

    //idx determines the place of the node on the base pyramid layer
		idx := task.Index % self.branches
    //lock the results for editing (NOTE: This may be too early?)
		results.Lock.Lock()
    //assign the node's place in the pyramid
		results.Levels[d][idx] = node
    //trigger the chunk channel 
		doChunk(chunkC, swg, hash, data, swg)

    //if there's only one chunk to be hashed (file size < chunk size) there's no more work to do
		if task.Index == 0 && last {
			pend.Done()
			break
		}

		//evaluate how many new parent :nodes are needed and where
		//on every idx==0 and after the very second node, we need to add a node in the parent layer of the pyramid
		if idx == 0 || task.Index == 1 {
			addNewLevel := false
			if float64(task.Index) == math.Pow(float64(self.branches), float64(results.Depth)) {
        //in this case, the pyramid grows an additional level
				m := make(map[int64]*Node)
				results.Levels = append(results.Levels, m)
				results.Depth++
				addNewLevel = true
			}
			j := results.Depth
			if addNewLevel == false {
				j--
			}
      //add a new node for every level needed, i.e:
      //[1][1], [1][2], [1][3], [2][0]
			for j > 0 {
				levelIndex := int64(len(results.Levels[j]))
        //just create a new node at that correspondent level, don't initialize children yet
				parent := &Node{0, 0, nil, last}
				results.Levels[j][levelIndex] = parent
				j--
			}
		}

    //for every node which completes the branch, or if it's the last node, build the nodes at the parent level
		if last || task.Index%self.branches == self.branches-1 {
			j := int64(1)
      //reset the hasher
			hasher.Reset()
      //iterate the complete tree - but...
			upper := results.Depth
      //...as long as we haven't processed the last node, we can't completely build the upper levels, only the parent
			if !last {
				upper = 1
			}
			for j <= upper {
				var end int64
        //the index of the node in the current level
				levelIndex := int64(len(results.Levels[j]) - 1)
        //parent
				pnode := results.Levels[j][levelIndex]
        //for the base layer pyramid, the start resets after the branch is full
				start := levelIndex / self.branches
        //at the base layer, we need to add the current chunk node's size 
				if j == 1 {
					pnode.Size = size
          //iterate until the last node in this branch
					end = start + task.Index%self.branches
				} else {
          //on upper levels, iterate all nodes of lower levels
					end = int64(start) + int64(len(results.Levels[j-1]))
				}
        //iterate nodes
				for k := start; k < end; {
					tmp := results.Levels[j-1][int64(k)]
					pnode.Size += tmp.Size
					pnode.Children = append(pnode.Children, tmp.Children[0])
					k++
          //if the current branch is full, iterate the next branch
					if k == self.branches-1 {
						levelIndex++
					}
				}
				if j == 1 {
          //if iterating chunk nodes, append the node as child to the parent
					pnode.Children = append(pnode.Children, node.Children[0])
				}
        //hash children
				data = make([]byte, hasher.Size()*len(pnode.Children)+8)
				binary.LittleEndian.PutUint64(data[:8], pnode.Size)

				hasher.Write(data[:8])
				for i, hash := range pnode.Children {
					copy(data[i*hasher.Size()+8:], hash[:])
					hasher.Write(hash[:])
				}
				bash := hasher.Sum(nil)
				copy(pnode.Children[0][:], bash)
				j++
        //trigger chunk channel
				doChunk(chunkC, swg, bash, data, swg)
			}
		}
    //NOTE: this may be too late
		results.Lock.Unlock()
		pend.Done()
	}

}

//"outsourced" to a func because couldn't figure out a more elegant way inside the processor function
func doChunk(chunkC chan *Chunk, swg *sync.WaitGroup, fKey Key, fData []byte, fwg *sync.WaitGroup) {
	if chunkC != nil {
		if swg != nil {
			swg.Add(1)
		}
		select {
		case chunkC <- &Chunk{Key: fKey, SData: fData, wg: fwg}:
			// case <- self.quitC
		}
	}

}
