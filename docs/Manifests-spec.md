---
tags: swarm, manifest
---

Swarm Manifests revisited
=============================




Let's define the entities involved:
* Manifest - the enclosing shell, root entry point, describes what we are describing, can specify access control,
* Index - a pluggable index (can be a compressed trie, htree, phtree, b-tree, b+tree etc)
* Inode - a filesystem inode representation, inherits ext filesystem attributes
* Feed - a swarm Feed 
* SwarmAC - a contract for selective disclosure of content between a publisher and granted parties

```

+------------------------------------+
|Manifest                            |
|------------------------------------|
|                                    |
| +--------------------------------+ |
| |Index                           | |
| |--------------------------------| |
| |                                | |
| | +------------+ +------------+  | |
| | |Inode       | |Feed        |  | |
| | |------------| |------------|  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | |            | |            |  | |
| | +------------+ +------------+  | |
| +--------------------------------+ |
| +--------------------------------+ |
| |AC                              | |
| |--------------------------------| |
| |                                | |
| |                                | |
| +--------------------------------+ |
+------------------------------------+
```

Our default index implementation Should satisfy `O(log n)` for add and update operation, i.e. cannot be a balanced tree. currently we will use the default compacted prefix trie implementation but in the future this could be practically any type of index

index properties:
```
type: "prefix" - for prefix trie
type: "flat" - no trie structure to reduce nesting
type: "btree"... etc
entries: array or any other data structure representation of the index
```
Let's define the manifest type and the operations on it. elements annotated with a '_' denote a value that has to be requested over the network, rather than local values that are already inlined inside the JSON file. this allows children to be optimized and inlined into the JSON file, rather than adding nesting levels and network complexity

```
type Manifest struct {
    type 	    string
    default     string //default path to resolve
    index       SwarmIndex //inlined index
    _index      string //swarm hash of index
    _ref        string //single entry manifest ref - in the case of act/feed/single hash pointer
    acl         SwarmACL
}

func (m *Manifest) Store //would store the manifest and the relevant index intermediate manifests should not be stored. this allows batch operations


type SwarmIndex interface {
	Get(key string) interface{}
	Put(key string, value interface{}) isNew bool
	Delete(key string) bool
}


type SwarmInode struct {
	ContentType string    `json:"contentType,omitempty"`
	Mode        int64     `json:"mode,omitempty"`
	Size        int64     `json:"size,omitempty"`
	ModTime     time.Time `json:"mod_time,omitempty"`
	Status      int       `json:"status,omitempty"`
}



```
The trie should define:
```
type Trie struct {
    key    string
    value	interface{}
    _value	string
    children	interface{} []trie
    _children	string
}


type Trier interface {
	Get(key string) interface{}
	Put(key string, value interface{}) isNew bool
	Delete(key string) bool
	Walk(walker WalkFunc) error
}
```


references:
* https://phunq.net/pipermail/tux3/2013-January/000026.html
* https://ext4.wiki.kernel.org/index.php/Ext4_Disk_Layout#Inode_Table
* http://ext2.sourceforge.net/2005-ols/paper-html/node3.html
