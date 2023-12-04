package crdt_go

import (
	_ "fmt"
	"github.com/google/uuid"
	"sort"
	"sync"
)

// BoundedPNCounter represents a positive-negative Counter CRDT.
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
func (c *BoundedPNCounter) Increment(NodeID string, amount uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if current, ok := c.positiveCount[NodeID]; ok {
		c.positiveCount[NodeID] = current + amount
	} else {
		c.positiveCount[NodeID] = amount
	}
}

// Decrement decrements the negative count for a given node.
func (c *BoundedPNCounter) Decrement(NodeID string, amount uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if current, ok := c.negativeCount[NodeID]; ok {
		if current+amount > c.positiveCount[NodeID] {
			c.negativeCount[NodeID] = c.positiveCount[NodeID]
		} else {
			c.negativeCount[NodeID] = current + amount
		}
	} else {
		c.negativeCount[NodeID] = amount
	}
}

// Value returns the computed value of the Counter.
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

	for NodeID, val1 := range c.positiveCount {
		val2, ok := other.positiveCount[NodeID]
		if !ok {
			continue
		}
		if val1 > val2 {
			return false
		}
	}

	for NodeID, val1 := range c.negativeCount {
		val2, ok := other.negativeCount[NodeID]
		if !ok {
			continue
		}
		if val1 > val2 {
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
	for NodeID, selfCount := range c.positiveCount {
		otherCount, ok := other.positiveCount[NodeID]
		if !ok {
			otherCount = 0
		}
		merged.positiveCount[NodeID] = max(selfCount, otherCount)
	}

	for NodeID, otherCount := range other.positiveCount {
		if _, ok := c.positiveCount[NodeID]; !ok {
			merged.positiveCount[NodeID] = max(0, otherCount)
		}
	}

	// Merge negative counts
	for NodeID, selfCount := range c.negativeCount {
		otherCount, ok := other.negativeCount[NodeID]
		if !ok {
			otherCount = 0
		}
		merged.negativeCount[NodeID] = max(selfCount, otherCount)
	}

	for NodeID, otherCount := range other.negativeCount {
		if _, ok := c.negativeCount[NodeID]; !ok {
			merged.negativeCount[NodeID] = max(0, otherCount)
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

// Clone creates a deep copy of the BoundedPNCounter.
func (c *BoundedPNCounter) Clone() *BoundedPNCounter {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new BoundedPNCounter with the same positive and negative counts
	clone := &BoundedPNCounter{
		positiveCount: make(map[string]uint32),
		negativeCount: make(map[string]uint32),
	}

	// Copy positive counts
	for NodeID, count := range c.positiveCount {
		clone.positiveCount[NodeID] = count
	}

	// Copy negative counts
	for NodeID, count := range c.negativeCount {
		clone.negativeCount[NodeID] = count
	}

	return clone
}

// AWSet represents an Add-Wins Set CRDT.
type AWSet struct {
	State   []item
	Context []ContextItem
}

type item struct {
	Name    string
	NodeID  string
	Counter uint32
}

type ContextItem struct {
	NodeID  string
	Counter uint32
}

// NewAWSet creates a new AWSet.
func NewAWSet() *AWSet {
	return &AWSet{
		State:   make([]item, 0),
		Context: make([]ContextItem, 0),
	}
}

func (s *AWSet) Clone() *AWSet {

	// Create a new AWSet and copy the exported fields
	clone := &AWSet{
		State:   make([]item, len(s.State)),
		Context: make([]ContextItem, len(s.Context)),
	}

	copy(clone.State, s.State)
	copy(clone.Context, s.Context)

	return clone
}

// Elements returns the unique elements in the AWSet.
func (s *AWSet) Elements() []string {

	uniqueItems := make(map[string]struct{})

	for _, item := range s.State {
		uniqueItems[item.Name] = struct{}{}
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

	for _, item := range s.State {
		if item.Name == itemName {
			return true
		}
	}

	return false
}

// MaxI returns the maximum Counter for a given node in the Context.
func (s *AWSet) MaxI(NodeID string) uint32 {

	var maxCounter uint32
	for _, ctxItem := range s.Context {
		if ctxItem.NodeID == NodeID && ctxItem.Counter > maxCounter {
			maxCounter = ctxItem.Counter
		}
	}

	return maxCounter
}

// NextI returns the next Counter for a given node.
func (s *AWSet) NextI(NodeID string) (string, uint32) {

	maxCounter := s.MaxI(NodeID) + 1
	return NodeID, maxCounter
}

// AddI adds an item to the AWSet.
func (s *AWSet) AddI(itemName string, NodeID string) {

	NodeID, Counter := s.NextI(NodeID)

	// Remove old Context items for the same node
	var newContext []ContextItem
	for _, ctxItem := range s.Context {
		if ctxItem.NodeID != NodeID {
			newContext = append(newContext, ctxItem)
		}
	}
	s.Context = newContext

	// Remove old State items with the same Name and node
	var newState []item
	for _, StateItem := range s.State {
		if StateItem.Name != itemName || StateItem.NodeID != NodeID {
			newState = append(newState, StateItem)
		}
	}
	s.State = newState

	// Add new Context and State items
	s.Context = append(s.Context, ContextItem{NodeID, Counter})
	s.State = append(s.State, item{itemName, NodeID, Counter})
}

// RmvI removes an item from the AWSet.
func (s *AWSet) RmvI(itemName string) {

	// Remove old State items with the same Name
	var newState []item
	for _, StateItem := range s.State {
		if StateItem.Name != itemName {
			newState = append(newState, StateItem)
		}
	}
	s.State = newState
}

// Filter returns the filtered State of the AWSet.
func (s *AWSet) Filter(incAWSet *AWSet) []item {

	var result []item

	for _, StateItem := range s.State {
		include := true
		for _, ctxItem := range incAWSet.Context {
			if StateItem.NodeID == ctxItem.NodeID && StateItem.Counter < ctxItem.Counter {
				include = false
				break
			}
		}

		if include {
			result = append(result, StateItem)
		}
	}

	return result
}

// exclusiveItemUnion returns the exclusive union of two slices of items.
func exclusiveItemUnion(slice1, slice2 []item) []item {
	set := make(map[item]struct{})

	// Add items from slice1 to the set
	for _, item := range slice1 {
		set[item] = struct{}{}
	}

	// Add items from slice2 to the set, excluding common elements
	for _, item := range slice2 {
		if _, exists := set[item]; !exists {
			set[item] = struct{}{}
		}
	}

	// Create a slice with the elements in the set
	result := make([]item, 0, len(set))
	for k := range set {
		result = append(result, k)
	}

	return result
}

// exclusiveContextItemUnion returns the exclusive union of two slices of ContextItems.
func exclusiveContextItemUnion(slice1, slice2 []ContextItem) []ContextItem {
	set := make(map[ContextItem]struct{})

	// Add items from slice1 to the set
	for _, item := range slice1 {
		set[item] = struct{}{}
	}

	// Add items from slice2 to the set, excluding common elements
	for _, item := range slice2 {
		if _, exists := set[item]; !exists {
			set[item] = struct{}{}
		}
	}

	// Create a slice with the elements in the set
	result := make([]ContextItem, 0, len(set))
	for k := range set {
		result = append(result, k)
	}

	return result
}

// Merge merges two AWSets.
func (s *AWSet) Merge(incAWSet *AWSet) {

	// Intersection of States
	var intersection []item
	for _, StateItem := range s.State {
		for _, incItem := range incAWSet.State {
			if StateItem.Name == incItem.Name && StateItem.NodeID == incItem.NodeID && StateItem.Counter == incItem.Counter {
				intersection = append(intersection, StateItem)
			}
		}
	}

	// Union of filtered States
	filteredState1 := s.Filter(incAWSet)
	filteredState2 := incAWSet.Filter(s)
	exclusiveUnionState := exclusiveItemUnion(filteredState1, filteredState2)
	exclusiveUnionState = exclusiveItemUnion(exclusiveUnionState, intersection)
	union := exclusiveUnionState

	// Union of Contexts
	exclusiveUnionContext := exclusiveContextItemUnion(s.Context, incAWSet.Context)
	unionContext := exclusiveUnionContext

	// Update AWSet
	s.State = union
	s.Context = unionContext
}

// ShoppingList represents a shopping list with CRDT support.
type ShoppingList struct {
	NodeID string
	items  map[string]*BoundedPNCounter
	awSet  *AWSet
}

// NewShoppingList creates a new ShoppingList.
func NewShoppingList() *ShoppingList {
	return &ShoppingList{
		NodeID: generateNodeID(),
		items:  make(map[string]*BoundedPNCounter),
		awSet:  NewAWSet(),
	}
}

// generateNodeID generates a unique node ID.
func generateNodeID() string {
	return uuid.New().String()
}

// AddOrUpdateItem adds or updates an item in the shopping list.
func (l *ShoppingList) AddOrUpdateItem(itemName string, quantityChange int) {

	if _, ok := l.items[itemName]; !ok {
		l.items[itemName] = NewBoundedPNCounter()
	}

	if quantityChange < 0 {
		l.items[itemName].Decrement(l.NodeID, uint32(-quantityChange))
		l.awSet.AddI(itemName, l.NodeID)
	} else if quantityChange > 0 {
		l.items[itemName].Increment(l.NodeID, uint32(quantityChange))
		l.awSet.AddI(itemName, l.NodeID)
	}
}

// RemoveItem removes an item from the shopping list.
func (l *ShoppingList) RemoveItem(itemName string) {

	l.awSet.RmvI(itemName)
	delete(l.items, itemName)
}

// Merge merges two shopping lists.
func (l *ShoppingList) Merge(incList *ShoppingList) {

	l.awSet.Merge(incList.awSet)

	// Merge items based on the merged AWSet
	for _, itemName := range l.awSet.Elements() {
		selfItem, selfExists := l.items[itemName]
		incItem, incExists := incList.items[itemName]

		if selfExists && incExists {
			l.items[itemName] = selfItem.Merge(incItem)
		} else if incExists {
			l.items[itemName] = incItem
		} else {
			// If the item only exists in one list, use that item
			l.items[itemName] = selfItem
		}
	}
}

// GetItems returns the Names of all items in the shopping list.
func (l *ShoppingList) GetItems() []string {

	return l.awSet.Elements()
}

// GetItemQuantity returns the quantity of an item in the shopping list. If the item does not exist, the second return is false.
func (l *ShoppingList) GetItemQuantity(itemName string) (int32, bool) {
	if !l.awSet.Contains(itemName) {
		return 0, false
	}
	return l.items[itemName].Value(), true
}

// Clone creates a deep copy of the ShoppingList.
func (l *ShoppingList) Clone() *ShoppingList {
	// Create a new ShoppingList with the same NodeID
	clone := &ShoppingList{
		NodeID: l.NodeID,
		items:  make(map[string]*BoundedPNCounter),
		awSet:  l.awSet.Clone(),
	}

	// Copy items from the original ShoppingList to the clone
	for itemName, counter := range l.items {
		clone.items[itemName] = counter.Clone()
	}

	return clone
}
