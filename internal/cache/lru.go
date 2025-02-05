package cache_lru

import (
	"fmt"
	"sync"
)

type ProxyItem struct {
	Status string
	Header string
	Body   []byte
}

type Node struct {
	Key      string
	Value    *ProxyItem
	NextNode *Node
	PrevNode *Node
}

type CacheList struct {
	sync.Mutex
	Size     int
	Capacity int
	Head     *Node
	Tail     *Node
	MyMap    map[string]*Node
}

func (ddl *CacheList) Get(key string) (bool, *ProxyItem) {
	ddl.Lock()
	defer ddl.Unlock()
	node, ok := ddl.MyMap[key]
	if !ok {
		return false, &ProxyItem{}
	}
	ddl.RemoveNode(node)
	ddl.UpdateMRU(node)

	return true, node.Value
}

func (ddl *CacheList) Put(key string, value *ProxyItem) {
	ddl.Lock()
	defer ddl.Unlock()
	/*
		if key exists: Update
	*/
	existingNode, ok := ddl.MyMap[key]
	if ok {
		// update the node in map and list
		existingNode.Value = value
		ddl.RemoveNode(existingNode)
		ddl.UpdateMRU(existingNode)
		ddl.MyMap[key] = existingNode
		return
	}

	// if key don't exist: Add (also update head and tail for the first times)
	node := Node{
		Key:   key,
		Value: value,
	}
	// create and insert into MyMap
	ddl.MyMap[key] = &node
	ddl.UpdateMRU(&node)
	ddl.Size++

	// if it is the first item
	if ddl.Tail == nil {
		ddl.Tail = &node
	}

	// evacuation
	if ddl.Size > ddl.Capacity {
		delete(ddl.MyMap, ddl.Tail.Key)
		ddl.RemoveNode(ddl.Tail)
		ddl.Size--
	}

}

func (ddl *CacheList) RemoveNode(node *Node) {
	if ddl.Head == node && ddl.Tail == node {
		ddl.Head = nil
		ddl.Tail = nil
	} else if node == ddl.Head {
		ddl.Head = node.NextNode
		ddl.Head.PrevNode = nil
	} else if node == ddl.Tail {
		ddl.Tail = node.PrevNode
		ddl.Tail.NextNode = nil
	} else {
		node.PrevNode.NextNode = node.NextNode
		node.NextNode.PrevNode = node.PrevNode
	}

	node.PrevNode = nil
	node.NextNode = nil
}

func (ddl *CacheList) UpdateMRU(node *Node) {
	// if already head, no need to update
	if ddl.Head == node {
		return
	}

	// if it is the first node
	if ddl.Head == nil {
		ddl.Head = node
		ddl.Tail = node
		node.PrevNode = nil
		node.NextNode = nil
		return
	}

	// insert
	node.NextNode = ddl.Head
	ddl.Head.PrevNode = node

	// update the nee Head
	ddl.Head = node
	node.PrevNode = nil
}

func (ddl *CacheList) DisplayOutput() {
	node := ddl.Head
	for node != nil {
		fmt.Print(node.Key, ":", node.Value, "->")
		node = node.NextNode
	}
	fmt.Println("END")
}
