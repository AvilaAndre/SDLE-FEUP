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
func (c *BoundedPNCounter) Value() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	var sumPositive, sumNegative uint32

	for _, v := range c.positiveCount {
		sumPositive += v
	}

	for _, v := range c.negativeCount {
		sumNegative += v
	}

	return int32(sumPositive - sumNegative)
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

func (c *BoundedPNCounter) Merge(other *BoundedPNCounter) *BoundedPNCounter {
	c.mu.Lock()
	defer c.mu.Unlock()

	merged := NewBoundedPNCounter()

	// Merge positive counts
	for nodeID, selfCount := range c.positiveCount {
		otherCount, ok := other.positiveCount[nodeID]
		if !ok {
			otherCount = 0
		}
		merged.positiveCount[nodeID] = max(selfCount, otherCount)
	}

	for nodeID, otherCount := range other.positiveCount {
		if _, ok := c.positiveCount[nodeID]; !ok {
			merged.positiveCount[nodeID] = max(0, otherCount)
		}
	}

	// Merge negative counts
	for nodeID, selfCount := range c.negativeCount {
		otherCount, ok := other.negativeCount[nodeID]
		if !ok {
			otherCount = 0
		}
		merged.negativeCount[nodeID] = max(selfCount, otherCount)
	}

	for nodeID, otherCount := range other.negativeCount {
		if _, ok := c.negativeCount[nodeID]; !ok {
			merged.negativeCount[nodeID] = max(0, otherCount)
		}
	}

	return merged
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
func (s *AWSet) Filter(incAWSet *AWSet) []item {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []item

	for _, stateItem := range s.state {
		include := true
		for _, ctxItem := range incAWSet.context {
			if stateItem.nodeID == ctxItem.nodeID && stateItem.counter < ctxItem.counter {
				include = false
				break
			}
		}

		if include {
			result = append(result, stateItem)
		}
	}

	return result
}

// Merge merges two AWSets.
func (s *AWSet) Merge(incAWSet *AWSet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Intersection of states
	var intersection []item
	for _, stateItem := range s.state {
		for _, incItem := range incAWSet.state {
			if stateItem.name == incItem.name && stateItem.nodeID == incItem.nodeID && stateItem.counter == incItem.counter {
				intersection = append(intersection, stateItem)
			}
		}
	}

	// Union of filtered states
	filteredState1 := s.Filter(incAWSet)
	filteredState2 := incAWSet.Filter(s)
	union := append(filteredState1, filteredState2...)
	union = append(union, intersection...)

	// Union of contexts
	unionContext := append(s.context, incAWSet.context...)

	// Update AWSet
	s.state = union
	s.context = unionContext
}

// ShoppingList represents a shopping list with CRDT support.
type ShoppingList struct {
	nodeID string
	items  map[string]*BoundedPNCounter
	awSet  *AWSet
	mu     sync.Mutex
}

// NewShoppingList creates a new ShoppingList.
func NewShoppingList() *ShoppingList {
	return &ShoppingList{
		nodeID: generateNodeID(),
		items:  make(map[string]*BoundedPNCounter),
		awSet:  NewAWSet(),
	}
}

// generateNodeID generates a unique node ID.
func generateNodeID() string {
	// Implement your logic to generate a unique node ID.
	return "unique_node_id"
}

// AddOrUpdateItem adds or updates an item in the shopping list.
func (l *ShoppingList) AddOrUpdateItem(itemName string, quantityChange int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.items[itemName]; !ok {
		l.items[itemName] = NewBoundedPNCounter()
	}

	if quantityChange < 0 {
		l.items[itemName].Decrement(l.nodeID, uint32(-quantityChange))
		l.awSet.AddI(itemName, l.nodeID)
	} else if quantityChange > 0 {
		l.items[itemName].Increment(l.nodeID, uint32(quantityChange))
		l.awSet.AddI(itemName, l.nodeID)
	}
}

// RemoveItem removes an item from the shopping list.
func (l *ShoppingList) RemoveItem(itemName string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.awSet.RmvI(itemName)
	delete(l.items, itemName)
}

// Merge merges two shopping lists.
func (l *ShoppingList) Merge(incList *ShoppingList) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.awSet.Merge(incList.awSet)

	// Merge items based on the merged AWSet
	for _, itemName := range l.awSet.Elements() {
		l.items[itemName] = l.items[itemName].Merge(incList.items[itemName])
	}
}

// GetItems returns the names of all items in the shopping list.
func (l *ShoppingList) GetItems() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.awSet.Elements()
}

// GetItemQuantity returns the quantity of an item in the shopping list.
func (l *ShoppingList) GetItemQuantity(itemName string) int32 {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.items[itemName].Value()
}
