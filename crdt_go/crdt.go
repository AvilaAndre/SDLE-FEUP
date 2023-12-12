package crdt_go

import (
	"encoding/json"
	"fmt"
	_ "fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// BoundedPNCounter represents a positive-negative Counter CRDT.
type BoundedPNCounter struct {
	PositiveCount map[string]uint32 `json:"positive_count"`
	NegativeCount map[string]uint32 `json:"negative_count"`
	mu            sync.Mutex
}

// NewBoundedPNCounter creates a new BoundedPNCounter.
func NewBoundedPNCounter() *BoundedPNCounter {
	return &BoundedPNCounter{
		PositiveCount: make(map[string]uint32),
		NegativeCount: make(map[string]uint32),
	}
}

// Increment increments the positive count for a given node.
func (c *BoundedPNCounter) Increment(NodeID string, amount uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if current, ok := c.PositiveCount[NodeID]; ok {
		c.PositiveCount[NodeID] = current + amount
	} else {
		c.PositiveCount[NodeID] = amount
	}
}

// Decrement decrements the negative count for a given node.
func (c *BoundedPNCounter) Decrement(NodeID string, amount uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if current, ok := c.NegativeCount[NodeID]; ok {
		if current+amount > c.PositiveCount[NodeID] {
			c.NegativeCount[NodeID] = c.PositiveCount[NodeID]
		} else {
			c.NegativeCount[NodeID] = current + amount
		}
	} else {
		c.NegativeCount[NodeID] = amount
	}
}

// Value returns the computed value of the Counter.
func (c *BoundedPNCounter) Value() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	var sumPositive, sumNegative uint32

	for _, v := range c.PositiveCount {
		sumPositive += v
	}

	for _, v := range c.NegativeCount {
		sumNegative += v
	}

	return int32(sumPositive - sumNegative)
}

// Compare compares two BoundedPNCounters.
func (c *BoundedPNCounter) Compare(other *BoundedPNCounter) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for NodeID, val1 := range c.PositiveCount {
		val2, ok := other.PositiveCount[NodeID]
		if !ok {
			continue
		}
		if val1 > val2 {
			return false
		}
	}

	for NodeID, val1 := range c.NegativeCount {
		val2, ok := other.NegativeCount[NodeID]
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
	for NodeID, selfCount := range c.PositiveCount {
		otherCount, ok := other.PositiveCount[NodeID]
		if !ok {
			otherCount = 0
		}
		merged.PositiveCount[NodeID] = max(selfCount, otherCount)
	}

	for NodeID, otherCount := range other.PositiveCount {
		if _, ok := c.PositiveCount[NodeID]; !ok {
			merged.PositiveCount[NodeID] = max(0, otherCount)
		}
	}

	// Merge negative counts
	for NodeID, selfCount := range c.NegativeCount {
		otherCount, ok := other.NegativeCount[NodeID]
		if !ok {
			otherCount = 0
		}
		merged.NegativeCount[NodeID] = max(selfCount, otherCount)
	}

	for NodeID, otherCount := range other.NegativeCount {
		if _, ok := c.NegativeCount[NodeID]; !ok {
			merged.NegativeCount[NodeID] = max(0, otherCount)
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
		PositiveCount: make(map[string]uint32),
		NegativeCount: make(map[string]uint32),
	}

	// Copy positive counts
	for NodeID, count := range c.PositiveCount {
		clone.PositiveCount[NodeID] = count
	}

	// Copy negative counts
	for NodeID, count := range c.NegativeCount {
		clone.NegativeCount[NodeID] = count
	}

	return clone
}

// AWSet represents an Add-Wins Set CRDT.
type AWSet struct {
	State   []item        `json:"state"`
	Context []ContextItem `json:"context"`
}

type item struct {
	Name    string
	NodeID  string
	Counter uint32
}

func (i item) MarshalJSON() ([]byte, error) {
	a := []interface{}{
		i.Name,
		i.NodeID,
		i.Counter,
	}

	return json.Marshal(a)
}

func (i *item) UnmarshalJSON(b []byte) error {
	var data []interface{}

	err := json.Unmarshal(b, &data)

	if err != nil {
		return err
	}

	if val, ok := data[0].(string); ok {
		i.Name = val
	} else {
		return fmt.Errorf("first value not of type string")
	}

	if val, ok := data[1].(string); ok {
		i.NodeID = val
	} else {
		return fmt.Errorf("second value not of type string")
	}

	if val, ok := data[2].(float64); ok {
		i.Counter = uint32(val)
	} else {
		return fmt.Errorf("third value not of type float")
	}

	return nil
}

type ContextItem struct {
	NodeID  string `json:"context"`
	Counter uint32 `json:"counter"`
}

func (ci ContextItem) MarshalJSON() ([]byte, error) {
	a := []interface{}{
		ci.NodeID,
		ci.Counter,
	}

	return json.Marshal(a)
}

func (ci *ContextItem) UnmarshalJSON(b []byte) error {
	var data []interface{}

	err := json.Unmarshal(b, &data)

	if err != nil {
		return err
	}

	if val, ok := data[0].(string); ok {
		ci.NodeID = val
	} else {
		return fmt.Errorf("first value not of type string")
	}

	if val, ok := data[1].(float64); ok {
		ci.Counter = uint32(val)
	} else {
		return fmt.Errorf("second value not of type float")
	}

	return nil
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
	NodeID string                       `json:"node_id"`
	Items  map[string]*BoundedPNCounter `json:"items"`
	AwSet  *AWSet                       `json:"awset"`
}

// NewShoppingList creates a new ShoppingList.
func NewShoppingList() *ShoppingList {
	return &ShoppingList{
		NodeID: generateNodeID(),
		Items:  make(map[string]*BoundedPNCounter),
		AwSet:  NewAWSet(),
	}
}

// generateNodeID generates a unique node ID.
func generateNodeID() string {
	return uuid.New().String()
}

// Use boolean and u32 like in the Rust version ?
// AddOrUpdateItem adds or updates an item in the shopping list.
func (l *ShoppingList) AddOrUpdateItem(itemName string, quantityChange int) {

	if _, ok := l.Items[itemName]; !ok {
		l.Items[itemName] = NewBoundedPNCounter()
	}

	if quantityChange < 0 {
		l.Items[itemName].Decrement(l.NodeID, uint32(-quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)
	} else if quantityChange > 0 {
		l.Items[itemName].Increment(l.NodeID, uint32(quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)
	}
}

// RemoveItem removes an item from the shopping list.
func (l *ShoppingList) RemoveItem(itemName string) {

	l.AwSet.RmvI(itemName)
	delete(l.Items, itemName)
}

// Merge merges two shopping lists.
func (l *ShoppingList) Merge(incList *ShoppingList) {

	l.AwSet.Merge(incList.AwSet)

	// Merge items based on the merged AWSet
	for _, itemName := range l.AwSet.Elements() {
		selfItem, selfExists := l.Items[itemName]
		incItem, incExists := incList.Items[itemName]

		if selfExists && incExists {
			l.Items[itemName] = selfItem.Merge(incItem)
		} else if incExists {
			l.Items[itemName] = incItem
		} else {
			// If the item only exists in one list, use that item
			l.Items[itemName] = selfItem
		}
	}
}

// GetItems returns the Names of all items in the shopping list.
func (l *ShoppingList) GetItems() []string {

	return l.AwSet.Elements()
}

// GetItemQuantity returns the quantity of an item in the shopping list. If the item does not exist, the second return is false.
func (l *ShoppingList) GetItemQuantity(itemName string) (int32, bool) {
	if !l.AwSet.Contains(itemName) {
		return 0, false
	}
	return l.Items[itemName].Value(), true
}

// Clone creates a deep copy of the ShoppingList.
func (l *ShoppingList) Clone() *ShoppingList {
	// Create a new ShoppingList with the same NodeID
	clone := &ShoppingList{
		NodeID: l.NodeID,
		Items:  make(map[string]*BoundedPNCounter),
		AwSet:  l.AwSet.Clone(),
	}

	// Copy items from the original ShoppingList to the clone
	for itemName, counter := range l.Items {
		clone.Items[itemName] = counter.Clone()
	}

	return clone
}

// ShoppingListV2 represents a shopping list with CRDT support.
type ShoppingListV2 struct {
	NodeID         string                       `json:"node_id"`
	NeededItems    map[string]*BoundedPNCounter `json:"needed"`
	PurchasedItems map[string]*BoundedPNCounter `json:"purchased"`
	AwSet          *AWSet                       `json:"awset"`
}

// NewShoppingListV2 creates a new ShoppingListV2.
func NewShoppingListV2() *ShoppingListV2 {
	return &ShoppingListV2{
		NodeID:         generateNodeID(),
		NeededItems:    make(map[string]*BoundedPNCounter),
		PurchasedItems: make(map[string]*BoundedPNCounter),
		AwSet:          NewAWSet(),
	}
}

// Use boolean and u32 like in the Rust version ?
// AddOrUpdateItem adds or updates an item in the shopping list.
func (l *ShoppingListV2) AddOrUpdateItem(itemName string, quantityChange int) {

	if _, ok := l.NeededItems[itemName]; !ok {
		l.NeededItems[itemName] = NewBoundedPNCounter()
	}

	if quantityChange < 0 {
		l.NeededItems[itemName].Decrement(l.NodeID, uint32(-quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)
	} else if quantityChange > 0 {
		l.NeededItems[itemName].Increment(l.NodeID, uint32(quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)
	}
}

// PurchaseItem adds or updates an item in the shopping list.
func (l *ShoppingListV2) PurchaseItem(itemName string, quantityChange int) {

	if _, ok := l.PurchasedItems[itemName]; !ok {
		l.PurchasedItems[itemName] = NewBoundedPNCounter()
	}

	if quantityChange < 0 {
		l.PurchasedItems[itemName].Decrement(l.NodeID, uint32(-quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)

	} else if quantityChange > 0 {
		l.PurchasedItems[itemName].Increment(l.NodeID, uint32(quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)

		l.NeededItems[itemName].Decrement(l.NodeID, uint32(quantityChange))
		l.AwSet.AddI(itemName, l.NodeID)
	}
}

// RemoveItem removes an item from the shopping list.
func (l *ShoppingListV2) RemoveItem(itemName string) {

	l.AwSet.RmvI(itemName)
	delete(l.NeededItems, itemName)
	delete(l.PurchasedItems, itemName)
}

// Merge merges two shopping lists.
func (l *ShoppingListV2) Merge(incList *ShoppingListV2) {

	l.AwSet.Merge(incList.AwSet)

	// Merge items based on the merged AWSet
	for _, itemName := range l.AwSet.Elements() {
		selfItem, selfExists := l.NeededItems[itemName]
		incItem, incExists := incList.NeededItems[itemName]

		if selfExists && incExists {
			l.NeededItems[itemName] = selfItem.Merge(incItem)
		} else if incExists {
			l.NeededItems[itemName] = incItem
		} else {
			// If the item only exists in one list, use that item
			l.NeededItems[itemName] = selfItem
		}
	}
	for _, itemName := range l.AwSet.Elements() {
		selfItem, selfExists := l.PurchasedItems[itemName]
		incItem, incExists := incList.PurchasedItems[itemName]

		if selfExists && incExists {
			l.PurchasedItems[itemName] = selfItem.Merge(incItem)
		} else if incExists {
			l.PurchasedItems[itemName] = incItem
		} else {
			// If the item only exists in one list, use that item
			l.PurchasedItems[itemName] = selfItem
		}
	}
}

// GetItems returns the Names of all items in the shopping list.
func (l *ShoppingListV2) GetItems() []string {

	return l.AwSet.Elements()
}

// GetItemQuantity returns the quantity of an item in the shopping list. If the item does not exist, the second return is false.
func (l *ShoppingListV2) GetItemQuantityNeeded(itemName string) (int32, bool) {
	if !l.AwSet.Contains(itemName) {
		return 0, false
	}
	return l.NeededItems[itemName].Value(), true
}

// GetItemQuantity returns the quantity of an item in the shopping list. If the item does not exist, the second return is false.
func (l *ShoppingListV2) GetItemQuantityPurchased(itemName string) (int32, bool) {
	if !l.AwSet.Contains(itemName) {
		return 0, false
	}
	return l.PurchasedItems[itemName].Value(), true
}

// Clone creates a deep copy of the ShoppingListV2.
func (l *ShoppingListV2) Clone() *ShoppingListV2 {
	// Create a new ShoppingListV2 with the same NodeID
	clone := &ShoppingListV2{
		NodeID:         l.NodeID,
		PurchasedItems: make(map[string]*BoundedPNCounter),
		NeededItems:    make(map[string]*BoundedPNCounter),
		AwSet:          l.AwSet.Clone(),
	}

	// Copy items from the original ShoppingListV2 to the clone
	for itemName, counter := range l.PurchasedItems {
		clone.PurchasedItems[itemName] = counter.Clone()
	}
	for itemName, counter := range l.NeededItems {
		clone.NeededItems[itemName] = counter.Clone()
	}

	return clone
}
