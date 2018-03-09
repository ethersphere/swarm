// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swarmdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"io"
	//"reflect"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	"strconv"
	"strings"
	"sync"
)

const binnum = 64
const STACK_SIZE = 100

type Val interface{}

type HashDB struct {
	rootnode   *Node
	swarmdb    *SwarmDB
	buffered   bool
	encrypted  int
	columnType sdbc.ColumnType
	mutex      sync.Mutex
}

type Node struct {
	Key        []byte
	Value      Val
	Next       bool
	Bin        []*Node
	Level      int
	Root       bool
	Version    int
	NodeKey    []byte //for disk/(net?)DB. Currently, it's bin data but it will be the hash
	NodeHash   []byte //for disk/(net?)DB. Currently, it's bin data but it will be the hash
	Loaded     bool
	Stored     bool
	columnType sdbc.ColumnType
	counter    int
}

type HashdbCursor struct {
	hashdb  *HashDB
	level   int
	bin     *stack_t
	node    *Node
	atlast  bool
	atfirst bool
}

// TODO: guarantee that this function will always work
func (self *HashDB) GetRootHash() []byte {
	return self.rootnode.NodeHash
}

func NewHashDB(u *SWARMDBUser, rootnode []byte, swarmdb *SwarmDB, columntype sdbc.ColumnType, encrypted int) (*HashDB, error) {
	hd := new(HashDB)
	n := NewNode(nil, nil)
	n.Root = true
	if rootnode == nil {
	} else {
		n.NodeHash = rootnode
		err := n.load(u, swarmdb, columntype)
		if err != nil {
			return nil, err
		}
	}
	hd.rootnode = n
	hd.swarmdb = swarmdb
	hd.buffered = false
	hd.encrypted = encrypted
	hd.columnType = columntype
	return hd, nil
}

func keyhash(k []byte) [32]byte {
	return sha3.Sum256(k)
}

func hashbin(k [32]byte, level int) int {
	x := 0x3F
	bytepos := level * 6 / 8
	bitpos := level * 6 % 8
	var fb int
	if bitpos <= 2 {
		fb = int(k[bytepos]) >> uint(2-bitpos)
	} else {
		fb = int(k[bytepos]) << uint(bitpos-2)
		fb = fb + (int(k[bytepos+1]) >> uint(8-(6-(8-bitpos))))
	}
	fb = fb & x
	return fb
}

func NewNode(k []byte, val Val) *Node {
	var nodelist = make([]*Node, binnum)
	var node = &Node{
		Key:      k,
		Next:     false,
		Bin:      nodelist,
		Value:    val,
		Level:    0,
		Root:     false,
		Version:  0,
		NodeKey:  nil,
		NodeHash: nil,
		Loaded:   false,
		Stored:   true,
	}
	return node
}

func NewRootNode(k []byte, val Val) *Node {
	return newRootNode(k, val, 0, 0, []byte("0:0"))
}

func newRootNode(k []byte, val Val, l int, version int, NodeKey []byte) *Node {
	var nodelist = make([]*Node, binnum)
	kh := keyhash(k)
	var bnum int
	bnum = hashbin(kh, l)
	newnodekey := string(NodeKey) + "|" + strconv.Itoa(bnum)
	var n = &Node{
		Key:     k,
		Next:    false,
		Bin:     nil,
		Value:   val,
		Level:   l + 1,
		Root:    false,
		Version: version,
		NodeKey: []byte(newnodekey),
	}

	nodelist[bnum] = n
	var rootnode = &Node{
		Key:     nil,
		Next:    true,
		Bin:     nodelist,
		Value:   nil,
		Level:   l,
		Root:    true,
		Version: version,
		NodeKey: NodeKey,
	}
	return rootnode
}

func (self *HashDB) Open(owner, tablename, columnname []byte) (bool, error) {
	return true, nil
}

func (self *HashDB) Put(u *SWARMDBUser, k []byte, v []byte) (bool, error) {
	err := self.rootnode.Add(u, k, v, self.swarmdb, self.columnType, self.encrypted)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (self *HashDB) GetRootNode() []byte {
	return self.rootnode.NodeHash
}

func (self *Node) Add(u *SWARMDBUser, k []byte, v Val, swarmdb *SwarmDB, columntype sdbc.ColumnType, encrypted int) error {
	log.Debug(fmt.Sprintf("HashDB Add ", self))
	self.Version++
	self.NodeKey = []byte("0")
	self.columnType = columntype
	_, err := self.add(u, NewNode(k, v), self.Version, self.NodeKey, swarmdb, columntype, encrypted)
	return err
}

func (self *Node) add(u *SWARMDBUser, addnode *Node, version int, nodekey []byte, swarmdb *SwarmDB, columntype sdbc.ColumnType, encrypted int) (newnode *Node, err error) {
	kh := keyhash(addnode.Key)
	bin := hashbin(kh, self.Level)
	self.NodeKey = nodekey
	self.Stored = false
	addnode.Stored = false
	addnode.columnType = columntype

	if self.Loaded == false {
		err = self.load(u, swarmdb, columntype)
		if err != nil {
			return nil, err
		}
		self.Loaded = true
	}

	if self.Next || self.Root {
		if self.Bin[bin] != nil {
			newnodekey := string(self.NodeKey) + "|" + strconv.Itoa(bin)
			if self.Bin[bin].Loaded == false {
				err := self.Bin[bin].load(u, swarmdb, columntype)
				if err != nil {
					return nil, err
				}
			}
			self.Bin[bin], err = self.Bin[bin].add(u, addnode, version, []byte(newnodekey), swarmdb, columntype, encrypted)
			if err != nil {
				return nil, err
			}
			var str string
			for i, b := range self.Bin {
				if b != nil {
					if b.Key != nil {
						str = str + "|" + strconv.Itoa(i) + ":" + string(b.Key)
					} else {
						str = str + "|" + strconv.Itoa(i)
					}
				}
			}
		} else {
			addnode.Level = self.Level + 1
			addnode.Loaded = true
			addnode.Stored = false
			addnode.Next = false
			addnode.NodeKey = []byte(string(self.NodeKey) + "|" + strconv.Itoa(bin))
			sdata := make([]byte, 4096)
			copy(sdata[64:], convertToByte(addnode.Value))
			copy(sdata[96:], addnode.Key)
			self.Bin[bin] = addnode
		}
	} else {
		if strings.Compare(string(self.Key), string(addnode.Key)) == 0 {
			sdata := make([]byte, 4096)
			copy(sdata[64:], convertToByte(addnode.Value))
			copy(sdata[96:], addnode.Key)
			dhash, err := swarmdb.StoreDBChunk(u, sdata, encrypted)
			if err != nil {
				return self, &sdbc.SWARMDBError{Message: `[hashdb:add] StoreDBChunk ` + err.Error()}
			}
			addnode.NodeHash = dhash
			self.Value = addnode.Value
			return self, nil
		}
		if len(self.Key) == 0 {
			// TODO: may be able to remove sdata
			sdata := make([]byte, 4096)
			copy(sdata[64:], convertToByte(addnode.Value))
			copy(sdata[96:], addnode.Key)
			addnode.Next = false
			addnode.Loaded = true
			self = addnode
			return self, nil
		}
		n := newRootNode(nil, nil, self.Level, version, self.NodeKey)
		n.Next = true
		n.Root = self.Root
		n.Level = self.Level
		n.Loaded = true
		addnode.Level = self.Level + 1
		cself := self
		cself.Level = self.Level + 1
		cself.Loaded = true
		n.add(u, addnode, version, self.NodeKey, swarmdb, columntype, encrypted)
		n.add(u, cself, version, self.NodeKey, swarmdb, columntype, encrypted)
		n.Loaded = true
		return n, nil
	}
	var svalue string
	for i, b := range self.Bin {
		if b != nil {
			svalue = svalue + "|" + strconv.Itoa(i)
		}
	}
	self.Loaded = true
	return self, nil
}

func compareVal(a, b Val) int {
	if va, ok := a.([]byte); ok {
		if vb, ok := b.([]byte); ok {
			return bytes.Compare(bytes.Trim(va, "\x00"), bytes.Trim(vb, "\x00"))
		}
	}
	return 100
}

func compareValType(a, b Val, columntype sdbc.ColumnType) int {
	if va, ok := a.([]byte); ok {
		if vb, ok := b.([]byte); ok {
			switch columntype {
			case sdbc.CT_INTEGER, sdbc.CT_FLOAT:
				for i := 0; i < 8; i++ {
					if va[i] > vb[i] {
						return 1
					} else if va[i] < vb[i] {
						return -1
					}
				}
				return 0
			default:
				return bytes.Compare(bytes.Trim(va, "\x00"), bytes.Trim(vb, "\x00"))
			}
		}
	}
	return 100
}

func convertToByte(a Val) []byte {
	if va, ok := a.([]byte); ok {
		return []byte(va)
	}
	if va, ok := a.(storage.Key); ok {
		return []byte(va)
	} else if va, ok := a.(string); ok {
		return []byte(va)
	}
	return nil
}

func (self *Node) storeBinToNetwork(u *SWARMDBUser, swarmdb *SwarmDB, encrypted int) ([]byte, error) {
	storedata := make([]byte, 66*64)

	if self.Next || self.Root {
		binary.LittleEndian.PutUint64(storedata[0:8], uint64(1))
	} else {
		binary.LittleEndian.PutUint64(storedata[0:8], uint64(0))
	}
	binary.LittleEndian.PutUint64(storedata[9:32], uint64(self.Level))

	for i, bin := range self.Bin {
		if bin != nil {
			copy(storedata[64+i*32:], bin.NodeHash)
		}
	}

	//wg := &sync.WaitGroup{}
	adhash, err := swarmdb.StoreDBChunk(u, storedata, encrypted)
	if err != nil {
		return adhash, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[hashdb:storeBinToNetwork] StoreDBChunk ", err.Error()))
	}
	//wg.Wait()
	return adhash, err
}

func (self *HashDB) Get(u *SWARMDBUser, k []byte) ([]byte, bool, error) {
	log.Debug("[hashdb:Get]")
	stack := newStack()
	ret, err := self.rootnode.Get(u, k, self.swarmdb, self.columnType, stack)
	if err != nil {
		switch err.(type) {
		case *sdbc.KeyNotFoundError:

			return nil, false, nil
		default:
			log.Debug(fmt.Sprintf("***** ERROR retrieving key [%s] ****** [%s]\n", k, err))
			return nil, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("Error Retrieving key [%s]", k))
		}
	}
	value := bytes.Trim(convertToByte(ret), "\x00")
	b := true
	if ret == nil {
		//var err sdbc.KeyNotFoundError
		//return nil, false, &err
		log.Debug("KEY NOT FOUND")
		return nil, false, nil
	}
	log.Debug(fmt.Sprintf("[hashdb:Get] Returning [%s]", value))
	return value, b, nil
}

func (self *HashDB) getStack(u *SWARMDBUser, k []byte) ([]byte, *stack_t, error) {
	stack := newStack()
	ret, err := self.rootnode.Get(u, k, self.swarmdb, self.columnType, stack)
	if err != nil {
		return nil, nil, err
	}
	value := bytes.Trim(convertToByte(ret), "\x00")
	if ret == nil {
		var err sdbc.KeyNotFoundError
		return nil, nil, &err
	}
	return value, stack, nil
}

func (self *Node) Get(u *SWARMDBUser, k []byte, swarmdb *SwarmDB, columntype sdbc.ColumnType, stack *stack_t) (Val, error) {
	kh := keyhash(k)
	bin := hashbin(kh, self.Level)

	if self.Loaded == false {
		err := self.load(u, swarmdb, columntype)
		if err != nil {
			return nil, err
		}
		self.Loaded = true
	}

	if self.Bin[bin] == nil {
		var err sdbc.KeyNotFoundError
		return nil, &err
	}
	if self.Bin[bin].Loaded == false {
		err := self.Bin[bin].load(u, swarmdb, columntype)
		if err != nil {
			//TODO: error check which error type
			return nil, err
		}
	}
	if self.Bin[bin].Next {
		stack.Push(bin)
		return self.Bin[bin].Get(u, k, swarmdb, columntype, stack)
	} else {
		if compareValType(k, self.Bin[bin].Key, columntype) == 0 && len(convertToByte(self.Bin[bin].Value)) > 0 {
			stack.Push(bin)
			return self.Bin[bin].Value, nil
		} else {
			//TODO: error check, no key error
			return nil, nil
		}
	}
	//TODO: error check, no key error
	return nil, nil
}

func (self *Node) load(u *SWARMDBUser, swarmdb *SwarmDB, columnType sdbc.ColumnType) error {
	buf, err := swarmdb.RetrieveDBChunk(u, self.NodeHash)
	if err != nil {
		return &sdbc.SWARMDBError{Message: `[hashdb:load] RetrieveDBChunk ` + err.Error()}
	}

	lf := int64(binary.LittleEndian.Uint64(buf[0:8]))
	if err != nil && err != io.EOF {
		//	fmt.Printf("\nError loading node: [%s]", err)
		self.Loaded = false
		self.Next = false
		return err
	}
	emptybyte := make([]byte, 32)
	if lf == 1 {
		for i := 0; i < 64; i++ {
			binnode := NewNode(nil, nil)
			binnode.NodeHash = make([]byte, 32)
			binnode.NodeHash = buf[64+32*i : 64+32*(i+1)]
			binnode.Loaded = false
			binnode.Level = self.Level + 1
			if binnode.NodeHash == nil || bytes.Compare(binnode.NodeHash, emptybyte) == 0 {
				self.Bin[i] = nil
			} else {
				self.Bin[i] = binnode
			}
		}
		self.Next = true
	} else {
		var pos int

		for pos = 96; pos < len(buf); pos++ {
			if buf[pos] == 0 {
				break
			}
		}
		if pos == 96 && bytes.Compare(buf[96:96+32], emptybyte) != 0 {
			pos = 96 + 32
		}
		if columnType == sdbc.CT_INTEGER {
			pos = 96 + 8
		}
		self.Key = buf[96:pos]
		self.Value = buf[64:96]
		self.Next = false
		if len(bytes.Trim(convertToByte(self.Value), "\x00")) == 0 {
			self.Key = nil
			self.Value = nil
			self.Loaded = true
			self.Next = false
			return nil
		}
	}
	self.Loaded = true
	return nil
}

func (self *HashDB) Insert(u *SWARMDBUser, k []byte, v []byte) (bool, error) {
	res, b, _ := self.Get(u, k)
	if res != nil || b {
		return false, &sdbc.SWARMDBError{Message: fmt.Sprintf(`[hashdb:Insert] Get - Key exists: %s`, string(k))}
	}
	_, err := self.Put(u, k, v)
	return true, err
}

func (self *HashDB) Delete(u *SWARMDBUser, k []byte) (bool, error) {
	_, b, err := self.rootnode.Delete(u, k, self.swarmdb, self.columnType)
	if err != nil {
		switch err.(type) {
		case *sdbc.KeyNotFoundError:
			return false, nil
		default:
			return false, err
		}
	}
	return b, nil
}

func (self *Node) Delete(u *SWARMDBUser, k []byte, swarmdb *SwarmDB, columntype sdbc.ColumnType) (newnode *Node, found bool, err error) {
	found = false
	if self.Loaded == false {
		err = self.load(u, swarmdb, columntype)
		if err != nil {
			return nil, false, err
		}
	}
	stack := newStack()
	ret, err := self.Get(u, k, swarmdb, columntype, stack)
	if ret == nil {
		return self, false, err
	}
	kh := keyhash(k)
	bin := hashbin(kh, self.Level)

	if self.Bin[bin] == nil {
		//TODO: need error??
		return nil, found, err
	}

	if self.Bin[bin].Next {
		self.Bin[bin], found, err = self.Bin[bin].Delete(u, k, swarmdb, columntype)
		if err != nil {
			return nil, false, err
		}
		if found {
			bincount := 0
			pos := -1
			for i, b := range self.Bin[bin].Bin {
				if b != nil {
					bincount++
					pos = i
				}
			}
			if bincount == 1 && self.Bin[bin].Bin[pos].Next == false {
				self.Bin[bin].Bin[pos].Level = self.Bin[bin].Level
				self.Bin[bin].Bin[pos] = self.Bin[bin].Bin[pos].shiftUpper()
				self.Bin[bin] = self.Bin[bin].Bin[pos]
			}
			self.Stored = false
			self.Bin[bin].Stored = false
		}
		return self, found, err
	} else {
		if self.Bin[bin].Loaded == false {
			self.Bin[bin].load(u, swarmdb, columntype)
		}
		if len(self.Bin[bin].Key) == 0 {
			return self, false, err
		}
		match := compareValType(k, self.Bin[bin].Key, columntype)
		if match != 0 {
			return self, found, err
		}
		self.Stored = false
		found = true
		self.Bin[bin] = nil
	}
	return self, found, err
}

func (self *Node) shiftUpper() *Node {
	for i, bin := range self.Bin {
		if bin != nil {
			if bin.Next == true {
				bin = bin.shiftUpper()
			}
			bin.Level = bin.Level - 1
			self.Bin[i] = bin
		}
	}
	return self
}

func (self *Node) Update(updatekey []byte, updatevalue []byte) (newnode *Node, err error) {
	kh := keyhash(updatekey)
	bin := hashbin(kh, self.Level)

	if self.Bin[bin] == nil {
		return self, &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:Update] No Key Error %x", updatekey)}
	}

	if self.Bin[bin].Next {
		return self.Bin[bin].Update(updatekey, updatevalue)
	} else {
		self.Bin[bin].Value = updatevalue
		return self, nil
	}
	return self, &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:Update] No Key Error %x", updatekey)}
}

func (self *HashDB) Close(u *SWARMDBUser) (bool, error) {
	return true, nil
}

func (self *HashDB) StartBuffer(u *SWARMDBUser) (bool, error) {
	self.buffered = true
	return true, nil
}

func (self *HashDB) FlushBuffer(u *SWARMDBUser) (bool, error) {
	if self.buffered == false {
		// do nothing: FlushBuffer does not require a StartBuffer
	}
	_, err := self.rootnode.flushBuffer(u, self.swarmdb, self.encrypted)
	if err != nil {
		return false, err
	}
	self.buffered = false
	return true, err
}

func (self *Node) flushBuffer(u *SWARMDBUser, swarmdb *SwarmDB, encrypted int) ([]byte, error) {
	var err error
	for _, bin := range self.Bin {
		if bin != nil {
			if bin.Next == true && bin.Stored == false {
				_, err := bin.flushBuffer(u, swarmdb, encrypted)
				if err != nil {
					return nil, err
				}
			} else if bin.Stored == false && len(bytes.Trim(convertToByte(bin.Value), "\x00")) > 0 {
				sdata := make([]byte, 4096)
				copy(sdata[64:], convertToByte(bin.Value))
				copy(sdata[96:], bin.Key)
				dhash, err := swarmdb.StoreDBChunk(u, sdata, encrypted)
				if err != nil {
					return nil, &sdbc.SWARMDBError{Message: `[hashdb:flushBuffer] StoreDBChunk ` + err.Error()}
				}
				bin.NodeHash = dhash
				bin.Stored = true
			}
		}
	}
	self.NodeHash, err = self.storeBinToNetwork(u, swarmdb, encrypted)
	self.Stored = true
	return self.NodeHash, err
}

func (self *HashDB) Print(u *SWARMDBUser) {
	self.rootnode.print(u, self.swarmdb, self.columnType)
	return
}

func (self *Node) print(u *SWARMDBUser, swarmdb *SwarmDB, columnType sdbc.ColumnType) {
	for binnum, bin := range self.Bin {
		if bin != nil {
			if bin.Loaded == false {
				bin.load(u, swarmdb, columnType)
				bin.Loaded = true
			}
			if bin.Next != true {
				fmt.Printf("leaf key = %v Value = %x binnum = %d level = %d Value len = %d\n", bin.Key, bin.Value, binnum, bin.Level, len(bytes.Trim(convertToByte(bin.Value), "\x00")))
			} else {
				fmt.Printf("node key = %v Value = %x binnum = %d level = %d\n", bin.Key, bin.Value, binnum, bin.Level)
				bin.print(u, swarmdb, columnType)
			}
		}
	}
}

func (self *HashDB) Seek(u *SWARMDBUser, k []byte) (OrderedDatabaseCursor, bool, error) {
	ret, stack, err := self.getStack(u, k)
	if err != nil {
		return nil, false, err
	}
	if ret == nil {
		return nil, false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:Seek] getStack - No Data")}
	}
	cursor, err := newHashdbCursor(self)
	if err != nil {
		return nil, false, err
	}
	node := self.rootnode
	for i := 0; i < stack.Size()-1; i++ {
		bin := stack.GetPos(i)
		node = node.Bin[bin]
	}
	cursor.bin = stack
	cursor.node = node
	cursor.level = stack.Size()
	return cursor, true, nil
}

func (self *HashDB) SeekFirst(u *SWARMDBUser) (OrderedDatabaseCursor, error) {
	cursor, err := newHashdbCursor(self)
	if err != nil {
		return nil, err
	}
	err = cursor.seeknext(u)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func (self *HashDB) SeekLast(u *SWARMDBUser) (OrderedDatabaseCursor, error) {
	cursor, err := newHashdbCursor(self)
	if err != nil {
		return nil, err
	}
	err = cursor.seekprev(u)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func newHashdbCursor(hashdb *HashDB) (*HashdbCursor, error) {
	cursor := &HashdbCursor{
		hashdb:  hashdb,
		level:   0,
		bin:     newStack(),
		node:    hashdb.rootnode,
		atlast:  false,
		atfirst: false,
	}
	return cursor, nil
}

func (self *HashdbCursor) Next(u *SWARMDBUser) ([]byte, []byte, error) {
	if self.atlast {
		return nil, nil, io.EOF
	}
	self.atfirst = false
	pos := self.bin.GetLast()
	k := convertToByte(self.node.Bin[pos].Key)
	v := bytes.Trim(convertToByte(self.node.Bin[pos].Value), "\x00")
	var err error
	if len(bytes.Trim(convertToByte(v), "\x00")) == 0 {
		err = self.seeknext(u)
		pos = self.bin.GetLast()
		k = convertToByte(self.node.Bin[pos].Key)
		v = convertToByte(self.node.Bin[pos].Value)
	}

	err = self.seeknext(u)
	if err != nil {
		if err == io.EOF {
			self.atlast = true
			err = nil
		}
		return k, v, err
	}
	if len(bytes.Trim(convertToByte(self.node.Bin[self.bin.GetLast()].Value), "\x00")) == 0 {
		err = self.seeknext(u)
		if err == io.EOF {
			self.atlast = true
			err = nil
		}
	}
	return k, v, err
}

func (self *HashdbCursor) Prev(u *SWARMDBUser) ([]byte, []byte, error) {
	if self.atfirst {
		return nil, nil, io.EOF
	}
	self.atlast = false
	pos := self.bin.GetLast()
	k := convertToByte(self.node.Bin[pos].Key)
	v := convertToByte(self.node.Bin[pos].Value)
	err := self.seekprev(u)
	if err != nil {
		if err == io.EOF {
			self.atfirst = true
			err = nil
		}
		return k, v, err
	}
	if len(bytes.Trim(convertToByte(self.node.Bin[self.bin.GetLast()].Value), "\x00")) == 0 {
		err = self.seekprev(u)
		if err == io.EOF {
			self.atfirst = true
			err = nil
		}
	}
	return k, v, err
}

// TODO: seek check it's needed or not
func (self *HashdbCursor) seek(u *SWARMDBUser, k []byte) error {
	return nil
}

func (self *HashdbCursor) seeknext(u *SWARMDBUser) error {
	l := self.level
	if self.node.Loaded == false {
		self.node.load(u, self.hashdb.swarmdb, self.hashdb.columnType)
	}

	lastpos := self.bin.GetLast()
	if lastpos < 0 {
		lastpos = 0
	} else {
		lastpos = lastpos + 1
	}
	for i := lastpos; i < 64; i++ {
		if self.node.Bin[i] != nil && self.node.Bin[i].Value != 0 {
			if self.node.Bin[i].Loaded == false {
				self.node.Bin[i].load(u, self.hashdb.swarmdb, self.hashdb.columnType)
			}
			if lastpos == 0 {
				self.level = l + 1
			}
			if self.node.Bin[i].Next == true {
				self.node = self.node.Bin[i]
				self.bin.Pop()
				self.bin.Push(i)
				self.bin.size = self.bin.size + 1
				if self.seeknext(u) == nil {
					return nil
				}
			} else {
				self.bin.Pop()
				self.bin.Push(i)
				return nil
			}
		}
	}
	if self.level == 0 {
		//TODO: check it's fine
		return io.EOF
	}
	self.level = self.level - 1
	bnum, err := self.bin.Pop()
	bnum, err = self.bin.Pop()
	if err != nil {
		//TODO: check it's fine
		return io.EOF
	}
	if bnum < 63 {
		self.bin.Push(bnum)
	} else {
		if self.bin.Size() == 0 {
			//TODO: check it's fine
			return io.EOF
		}
		bnum, _ := self.bin.Pop()
		self.bin.Push(bnum + 1)
		self.level = self.level - 1
	}
	self.node = self.hashdb.rootnode
	for i := 0; i < self.bin.Size()-1; i++ {
		if self.bin.GetPos(i) == -1 {
			return &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:seeknext] No Data")}
		}
		if self.node.Bin[self.bin.GetPos(i)] == nil {
		} else {
			if self.node.Bin[self.bin.GetPos(i)].Loaded == false {
				self.node.Bin[self.bin.GetPos(i)].load(u, self.hashdb.swarmdb, self.hashdb.columnType)
			}
			self.node = self.node.Bin[self.bin.GetPos(i)]
			//return nil
		}
	}
	err = self.seeknext(u)
	return err
}

func (self *HashdbCursor) seekprev(u *SWARMDBUser) error {
	l := self.level
	if self.node.Loaded == false {
		self.node.load(u, self.hashdb.swarmdb, self.hashdb.columnType)
	}

	lastpos := self.bin.GetLast()
	if lastpos < 0 {
		lastpos = 63
	} else if lastpos == 0 {
		lastpos = 63
	} else {
		lastpos = lastpos - 1
	}
	for i := lastpos; i >= 0; i-- {
		if self.node.Bin[i] != nil && self.node.Bin[i].Value != 0 {
			if self.node.Bin[i].Loaded == false {
				self.node.Bin[i].load(u, self.hashdb.swarmdb, self.hashdb.columnType)
			}
			self.level = l + 1
			if self.node.Bin[i].Next == true {
				self.node = self.node.Bin[i]
				self.bin.Pop()
				self.bin.Push(i)
				self.bin.size = self.bin.size + 1
				if self.seekprev(u) == nil {
					return nil
				}
			} else {
				self.bin.Pop()
				self.bin.Push(i)
				return nil
			}
		}
	}
	self.bin.Pop()
	if self.level == 0 {
		//TODO: check it's okay
		return io.EOF
	}
	self.level = self.level - 1
	bnum, err := self.bin.Pop()
	if err != nil {
		//TODO: check it's okay
		return io.EOF
	}

	if bnum != 0 {
		self.bin.Push(bnum)
	} else {
		if self.bin.Size() == 0 {
			//TODO: check it's okay
			return io.EOF
		}
		bnum, _ := self.bin.Pop()
		self.bin.Push(bnum - 1)
		self.level = self.level - 1
	}
	self.node = self.hashdb.rootnode
	for i := 0; i < self.bin.Size()-1; i++ {
		if self.bin.GetPos(i) == -1 {
			return &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:seekprev] No data")}
		}
		if self.node.Bin[self.bin.GetPos(i)] == nil {
		} else {
			if self.node.Bin[self.bin.GetPos(i)].Loaded == false {
				self.node.Bin[self.bin.GetPos(i)].load(u, self.hashdb.swarmdb, self.hashdb.columnType)
			}
			self.node = self.node.Bin[self.bin.GetPos(i)]
		}
	}
	return self.seekprev(u)
}

//TODO:: check it's needed
func (self *HashdbCursor) seeklast() error {
	return nil
}

type stack_t struct {
	data []int
	size int
}

func newStack() *stack_t {
	s := stack_t{
		data: make([]int, STACK_SIZE),
		size: 0,
	}
	for i := 0; i < STACK_SIZE; i++ {
		s.data[i] = -1
	}
	return &s
}

func (self *stack_t) Push(add int) error {
	if self.size+1 > STACK_SIZE {
		return &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:Push] over max stack")}
	}
	self.data[self.size] = add
	self.size = self.size + 1
	return nil
}

func (self *stack_t) Pop() (int, error) {
	if self.size == 0 {
		return -1, &sdbc.SWARMDBError{Message: fmt.Sprintf("[hashdb:Pop] nothing in stack")}
	}
	pos := self.data[self.size-1]
	self.data[self.size-1] = -1
	self.size = self.size - 1
	return pos, nil
}

func (self *stack_t) GetLast() int {
	if self.size <= 0 {
		return -1
	}
	return self.data[self.size-1]
}

func (self *stack_t) GetFirst() int {
	return self.data[0]
}

func (self *stack_t) GetPos(pos int) int {
	if self.size < pos {
		return -1
	}
	return self.data[pos]
}

func (self *stack_t) Size() int {
	return self.size
}
