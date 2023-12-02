package crdt_go

import (
	_ "fmt"
	"sort"
	"sync"
)

// BoundedPNCounter represents a positive-negative counter CRDT.
type BoundedPNCounter struct {
	positiveCount map[string]uint32
	negativeCount map[string]uint32
	mu            sync.Mutex
}

// NewBoundedPNCounter creates a new BoundedPNCounter.
func NewBoundedPNCounter() *BoundedPNCounter {
	return &BoundedPNCounter{
		positiveCount: make(map[string]uint32),
		negativeCount: make(map[string]uint32),
	}
}

// Increment increments the positive count for a given node.
func (c *BoundedPNCounter) Increment(nodeID string, amount uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if current, ok := c.positiveCount[nodeID]; ok {
		c.positiveCount[nodeID] = current + amount
	} else {
		c.positiveCount[nodeID] = amount
	}
}

// Decrement decrements the negative count for a given node.
func (c *BoundedPNCounter) Decrement(nodeID string, amount uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if current, ok := c.negativeCount[nodeID]; ok {
		if current+amount > c.positiveCount[nodeID] {
			c.negativeCount[nodeID] = c.positiveCount[nodeID]
		} else {
			c.negativeCount[nodeID] = current + amount
		}
	} else {
		c.negativeCount[nodeID] = amount
	}
}

// Value returns the computed value of the counter.
func (c *BoundedPNCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var sumPositive, sumNegative uint32

	for _, v := range c.positiveCount {
		sumPositive += v
	}

	for _, v := range c.negativeCount {
		sumNegative += v
	}

	return int(sumPositive - sumNegative)
}

// Compare compares two BoundedPNCounters.
func (c *BoundedPNCounter) Compare(other *BoundedPNCounter) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for nodeID, val1 := range c.positiveCount {
		val2, ok := other.positiveCount[nodeID]
		if !ok || val1 > val2 {
			return false
		}
	}

	for nodeID, val1 := range c.negativeCount {
		val2, ok := other.negativeCount[nodeID]
		if !ok || val1 > val2 {
			return false
		}
	}

	return true
}

// Merge merges two BoundedPNCounters.
func (c *BoundedPNCounter) Merge(other *BoundedPNCounter) *BoundedPNCounter {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := NewBoundedPNCounter()

	for nodeID, val1 := range c.positiveCount {
		val2, _ := other.positiveCount[nodeID]
		result.positiveCount[nodeID] = max(val1, val2)
	}

	for nodeID, val1 := range c.negativeCount {
		val2, _ := other.negativeCount[nodeID]
		result.negativeCount[nodeID] = max(val1, val2)
	}

	return result
}

// max returns the maximum of two uint32 values.
func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// AWSet represents an Add-Wins Set CRDT.
type AWSet struct {
	state   []item
	context []contextItem
	mu      sync.Mutex
}

type item struct {
	name    string
	nodeID  string
	counter uint32
}

type contextItem struct {
	nodeID  string
	counter uint32
}

// NewAWSet creates a new AWSet.
func NewAWSet() *AWSet {
	return &AWSet{
		state:   make([]item, 0),
		context: make([]contextItem, 0),
	}
}

// Elements returns the unique elements in the AWSet.
func (s *AWSet) Elements() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	uniqueItems := make(map[string]struct{})

	for _, item := range s.state {
		uniqueItems[item.name] = struct{}{}
	}

	result := make([]string, 0, len(uniqueItems))
	for item := range uniqueItems {
		result = append(result, item)
	}

	sort.Strings(result)
	return result
}

// Contains checks if the given element is in the AWSet.
func (s *AWSet) Contains(itemName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.state {
		if item.name == itemName {
			return true
		}
	}

	return false
}

// MaxI returns the maximum counter for a given node in the context.
func (s *AWSet) MaxI(nodeID string) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var maxCounter uint32
	for _, ctxItem := range s.context {
		if ctxItem.nodeID == nodeID && ctxItem.counter > maxCounter {
			maxCounter = ctxItem.counter
		}
	}

	return maxCounter
}

// NextI returns the next counter for a given node.
func (s *AWSet) NextI(nodeID string) (string, uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxCounter := s.MaxI(nodeID) + 1
	return nodeID, maxCounter
}

// AddI adds an item to the AWSet.
func (s *AWSet) AddI(itemName string, nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeID, counter := s.NextI(nodeID)

	// Remove old context items for the same node
	var newContext []contextItem
	for _, ctxItem := range s.context {
		if ctxItem.nodeID != nodeID {
			newContext = append(newContext, ctxItem)
		}
	}
	s.context = newContext

	// Remove old state items with the same name and node
	var newState []item
	for _, stateItem := range s.state {
		if stateItem.name != itemName || stateItem.nodeID != nodeID {
			newState = append(newState, stateItem)
		}
	}
	s.state = newState

	// Add new context and state items
	s.context = append(s.context, contextItem{nodeID, counter})
	s.state = append(s.state, item{itemName, nodeID, counter})
}

// RmvI removes an item from the AWSet.
func (s *AWSet) RmvI(itemName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove old state items with the same name
	var newState []item
	for _, stateItem := range s.state {
		if stateItem.name != itemName {
			newState = append(newState, stateItem)
		}
	}
	s.state = newState
}

// Filter returns the filtered state of the AWSet.
func (s *AWSet) Filter(include func(itemName string) bool) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []string
	for _, item := range s.state {
		if include(item.name) {
			result = append(result, item.name)
		}
	}

	sort.Strings(result)
	return result
}
