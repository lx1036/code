package skiplist

import (
	"errors"
	"math/rand"
	"sync"
)

type ConcurrentSkipList struct {
	skipLists []*skipList
	level     int
}

type skipList struct {
	level  int
	length int32
	head   *Node
	tail   *Node
	mutex  sync.RWMutex
}

type Node struct {
	index     uint64
	value     interface{}
	nextNodes []*Node //
}

const (
	SHARDS      = 32
	MaxLevel    = 32
	PROBABILITY = 0.25
)

func (list *ConcurrentSkipList) Level() int {
	return list.level
}

func (list *ConcurrentSkipList) Length() int32 {
	var length int32
	for _, skip := range list.skipLists {
		length += skip.getLength()
	}

	return length
}

/**

 */
func (list *ConcurrentSkipList) Insert(index uint64, value interface{}) {
	skipList := list.skipLists[index]
	skipList.insert(index, value)
}

func NewConcurrentSkipList(level int) (*ConcurrentSkipList, error) {
	if level <= 0 || level > MaxLevel {
		return nil, errors.New("level must between 1 and 32")
	}

	skipLists := make([]*skipList, SHARDS, SHARDS)
	for i := 0; i < SHARDS; i++ {
		skipLists[i] = newSkipList(level)
	}

	return &ConcurrentSkipList{
		skipLists: skipLists,
		level:     level,
	}, nil
}

func (list *skipList) getLength() int32 {
	return list.length
}

/**

 */
func (list *skipList) insert(index uint64, value interface{}) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	previousNodes, currentNode := list.searchWithPreviousNodes(index)

	newNode := newNode(index, value, list.randomLevel())

	for i := len(newNode.nextNodes) - 1; i >= 0; i-- {
		newNode.nextNodes[i] = previousNodes[i].nextNodes[i]
	}

}

func (list *skipList) randomLevel() int {
	level := 1
	for rand.Float64() < PROBABILITY && level < list.level {
		level++
	}

	return level
}

func (list *skipList) searchWithPreviousNodes(index uint64) ([]*Node, *Node) {
	currentNode := list.head
	previousNodes := make([]*Node, list.level)

	for level := list.level - 1; level >= 0; level-- {
		if currentNode.nextNodes[level] != list.tail && currentNode.nextNodes[level].index < index {
			currentNode = currentNode.nextNodes[level]
		}

		previousNodes[level] = currentNode
	}

	return previousNodes, currentNode

}

func newNode(index uint64, value interface{}, level int) *Node {
	return &Node{
		index:     index,
		value:     value,
		nextNodes: make([]*Node, level, level),
	}
}

func newSkipList(level int) *skipList {
	head := newNode(0, nil, level)
	var tail *Node
	for i := 0; i < len(head.nextNodes); i++ {
		head.nextNodes[i] = tail
	}

	return &skipList{
		level:  level,
		length: 0,
		head:   head,
		tail:   tail,
		mutex:  sync.RWMutex{},
	}
}
