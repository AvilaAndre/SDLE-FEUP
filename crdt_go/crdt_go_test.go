package crdt_go

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var numberOfProperties = 100000

func TestBoundedPNCounter(t *testing.T) {
	// Create two BoundedPNCounters
	counter1 := NewBoundedPNCounter()
	counter2 := NewBoundedPNCounter()

	// Increment and decrement values
	counter1.Increment("node1", 3)
	counter2.Decrement("node2", 2)

	// Test Value function
	if value := counter1.Value(); value != 3 {
		t.Errorf("Expected value to be 3, got %d", value)
	}

	// Test Compare function
	if !counter1.Compare(counter2) {
		t.Error("Expected counters to be equal")
	}

	// Test Merge function
	expectedMergedValue := counter1.Value() + counter2.Value()
	mergedCounter := counter1.Merge(counter2)

	if mergedCounter.Value() != expectedMergedValue {
		t.Errorf("Expected merged value to be %d, got %d", expectedMergedValue, mergedCounter.Value())
	}
}

func TestIncrementBoundedPNCounter(t *testing.T) {
	c := NewBoundedPNCounter()
	nodeID := "test_node"
	c.Increment(nodeID, 5)
	if count, ok := c.PositiveCount[nodeID]; !ok || count != 5 {
		t.Errorf("Increment failed. Expected positiveCount[%s] to be 5, got %v", nodeID, count)
	}
}

func TestDecrementBoundedPNCounter(t *testing.T) {
	c := NewBoundedPNCounter()
	nodeID := "test_node"
	c.Increment(nodeID, 5)
	c.Decrement(nodeID, 3)
	if count, ok := c.NegativeCount[nodeID]; !ok || count != 3 {
		t.Errorf("Decrement failed. Expected NegativeCount[%s] to be 3, got %v", nodeID, count)
	}
}

func TestValueBoundedPNCounter(t *testing.T) {
	c := NewBoundedPNCounter()
	nodeID := "test_node"
	c.Increment(nodeID, 10)
	c.Decrement(nodeID, 4)
	if value := c.Value(); value != 6 {
		t.Errorf("Value failed. Expected Value() to be 6, got %v", value)
	}
}

func TestLowerBoundaryBoundedPNCounter(t *testing.T) {
	c := NewBoundedPNCounter()
	nodeID := "test_node"
	c.Increment(nodeID, 10)
	c.Decrement(nodeID, 4)
	c.Decrement(nodeID, 10) // Assuming decrement is bounded by positive count
	if value := c.Value(); value != 0 {
		t.Errorf("LowerBoundary failed. Expected Value() to be 0, got %v", value)
	}
}

func TestCompareBoundedPNCounter(t *testing.T) {
	c1 := NewBoundedPNCounter()
	c2 := NewBoundedPNCounter()
	nodeID := "test_node"
	c1.Increment(nodeID, 3)
	c2.Increment(nodeID, 5)
	if !c1.Compare(c2) {
		t.Errorf("Compare failed. Expected c1 to be less than c2")
	}
}

func TestMergeSameKeysBoundedPNCounter(t *testing.T) {
	c1 := NewBoundedPNCounter()
	c2 := NewBoundedPNCounter()
	nodeID := "test_node"
	c1.Increment(nodeID, 2)
	c2.Increment(nodeID, 3)
	merged := c1.Merge(c2)
	if count, ok := merged.PositiveCount[nodeID]; !ok || count != 3 {
		t.Errorf("MergeSameKeys failed. Expected positiveCount[%s] to be 3, got %v", nodeID, count)
	}
}

func TestMergeDisjointKeysBoundedPNCounter(t *testing.T) {
	c1 := NewBoundedPNCounter()
	c2 := NewBoundedPNCounter()
	node1 := "node1"
	node2 := "node2"
	c1.Increment(node1, 2)
	c2.Increment(node2, 3)
	merged := c1.Merge(c2)
	if count1, ok := merged.PositiveCount[node1]; !ok || count1 != 2 {
		t.Errorf("MergeDisjointKeys failed. Expected positiveCount[%s] to be 2, got %v", node1, count1)
	}
	if count2, ok := merged.PositiveCount[node2]; !ok || count2 != 3 {
		t.Errorf("MergeDisjointKeys failed. Expected positiveCount[%s] to be 3, got %v", node2, count2)
	}
}

func TestMergeEmptyCountersBoundedPNCounter(t *testing.T) {
	c1 := NewBoundedPNCounter()
	c2 := NewBoundedPNCounter()
	merged := c1.Merge(c2)
	if len(merged.PositiveCount) != 0 || len(merged.NegativeCount) != 0 {
		t.Errorf("MergeEmptyCounters failed. Expected both positiveCount and negativeCount to be empty.")
	}
}

func TestMergeOneEmptyCounterBoundedPNCounter(t *testing.T) {
	c1 := NewBoundedPNCounter()
	c2 := NewBoundedPNCounter()
	nodeID := "test_node"
	c1.Increment(nodeID, 1)
	merged := c1.Merge(c2)
	if count, ok := merged.PositiveCount[nodeID]; !ok || count != 1 {
		t.Errorf("MergeOneEmptyCounter failed. Expected positiveCount[%s] to be 1, got %v", nodeID, count)
	}
}

func TestAWSet(t *testing.T) {
	// Create two AWSets
	set1 := NewAWSet()
	set2 := NewAWSet()

	// Add items to the sets
	set1.AddI("item1", "node1")
	set2.AddI("item2", "node2")

	// Test Elements function
	elements1 := set1.Elements()
	elements2 := set2.Elements()
	if len(elements1) != 1 || elements1[0] != "item1" {
		t.Errorf("Expected elements1 to be ['item1'], got %v", elements1)
	}
	if len(elements2) != 1 || elements2[0] != "item2" {
		t.Errorf("Expected elements2 to be ['item2'], got %v", elements2)
	}

	// Test Contains function
	if !set1.Contains("item1") {
		t.Error("Expected set1 to contain 'item1'")
	}
	if set1.Contains("item2") {
		t.Error("Expected set1 not to contain 'item2'")
	}

	// Test Merge function
	set1.Merge(set2)
	elementsMerged := set1.Elements()
	if len(elementsMerged) != 2 || !(elementsMerged[0] == "item1" && elementsMerged[1] == "item2") {
		t.Errorf("Expected merged elements to be ['item1', 'item2'], got %v", elementsMerged)
	}
}

func TestAWSetNew(t *testing.T) {
	awset := NewAWSet()
	if len(awset.State) != 0 || len(awset.Context) != 0 {
		t.Errorf("NewAWSet() failed, expected empty State and Context")
	}
}

func TestMaxI(t *testing.T) {
	awset := NewAWSet()
	nodeID := "test_node"
	awset.Context = append(awset.Context, ContextItem{nodeID, 1})
	awset.Context = append(awset.Context, ContextItem{nodeID, 3})
	awset.Context = append(awset.Context, ContextItem{nodeID, 2})

	maxCounter := awset.MaxI(nodeID)
	if maxCounter != 3 {
		t.Errorf("MaxI() failed, expected 3, got %v", maxCounter)
	}
}

func TestNextI(t *testing.T) {
	awset := NewAWSet()
	nodeID := "test_node"
	awset.Context = append(awset.Context, ContextItem{nodeID, 1})
	awset.Context = append(awset.Context, ContextItem{nodeID, 2})

	nextNodeID, nextCounter := awset.NextI(nodeID)
	if nextNodeID != nodeID || nextCounter != 3 {
		t.Errorf("NextI() failed, expected (%s, 3), got (%s, %v)", nodeID, nextNodeID, nextCounter)
	}
}

func TestContextWithMultipleNodes(t *testing.T) {
	awset := NewAWSet()
	nodeID1 := "node1"
	nodeID2 := "node2"

	awset.Context = append(awset.Context, ContextItem{nodeID1, 1})
	awset.Context = append(awset.Context, ContextItem{nodeID1, 2})
	awset.Context = append(awset.Context, ContextItem{nodeID2, 1})
	awset.Context = append(awset.Context, ContextItem{nodeID2, 3})

	if maxCounter := awset.MaxI(nodeID1); maxCounter != 2 {
		t.Errorf("MaxI() failed for %s, expected 2, got %v", nodeID1, maxCounter)
	}

	if maxCounter := awset.MaxI(nodeID2); maxCounter != 3 {
		t.Errorf("MaxI() failed for %s, expected 3, got %v", nodeID2, maxCounter)
	}

	nextNodeID, nextCounter := awset.NextI(nodeID1)
	if nextNodeID != nodeID1 || nextCounter != 3 {
		t.Errorf("NextI() failed for %s, expected (%s, 3), got (%s, %v)", nodeID1, nodeID1, nextNodeID, nextCounter)
	}

	nextNodeID, nextCounter = awset.NextI(nodeID2)
	if nextNodeID != nodeID2 || nextCounter != 4 {
		t.Errorf("NextI() failed for %s, expected (%s, 4), got (%s, %v)", nodeID2, nodeID2, nextNodeID, nextCounter)
	}
}

func TestAddNewItem(t *testing.T) {
	awset := NewAWSet()
	nodeID := "test_node"
	itemName := "apple"

	awset.AddI(itemName, nodeID)

	if !containsItem(awset.State, itemName, nodeID, 1) {
		t.Errorf("AddI() failed, expected State to contain (%s, %s, 1)", itemName, nodeID)
	}

	if !containsContext(awset.Context, nodeID, 1) {
		t.Errorf("AddI() failed, expected Context to contain (%s, 1)", nodeID)
	}
}

func TestIncrementExistingItem(t *testing.T) {
	awset := NewAWSet()
	nodeID := "test_node"
	itemName := "apple"
	awset.State = append(awset.State, item{itemName, nodeID, 1})
	awset.Context = append(awset.Context, ContextItem{nodeID, 1})

	awset.AddI(itemName, nodeID)

	if !containsItem(awset.State, itemName, nodeID, 2) {
		t.Errorf("AddI() failed, expected State to contain (%s, %s, 2)", itemName, nodeID)
	}

	if !containsContext(awset.Context, nodeID, 2) {
		t.Errorf("AddI() failed, expected Context to contain (%s, 2)", nodeID)
	}
}

func TestDecrementExistingItem(t *testing.T) {
	awset := NewAWSet()
	nodeID := "test_node"
	itemName := "apple"
	awset.State = append(awset.State, item{itemName, nodeID, 1})
	awset.Context = append(awset.Context, ContextItem{nodeID, 1})

	awset.AddI(itemName, nodeID)

	if !containsItem(awset.State, itemName, nodeID, 2) {
		t.Errorf("AddI() failed, expected State to contain (%s, %s, 2)", itemName, nodeID)
	}

	if !containsContext(awset.Context, nodeID, 2) {
		t.Errorf("AddI() failed, expected Context to contain (%s, 2)", nodeID)
	}
}

func TestAddI(t *testing.T) {
	awset := NewAWSet()
	nodeID := "test_node"
	itemName := "apple"

	awset.AddI(itemName, nodeID)

	if !containsItem(awset.State, itemName, nodeID, 1) {
		t.Errorf("AddI() failed, expected State to contain (%s, %s, 1)", itemName, nodeID)
	}

	if !containsContext(awset.Context, nodeID, 1) {
		t.Errorf("AddI() failed, expected Context to contain (%s, 1)", nodeID)
	}
}

func containsItem(State []item, name string, nodeID string, counter uint32) bool {
	for _, i := range State {
		if i.Name == name && i.NodeID == nodeID && i.Counter == counter {
			return true
		}
	}
	return false
}

func containsContext(Context []ContextItem, nodeID string, counter uint32) bool {
	for _, c := range Context {
		if c.NodeID == nodeID && c.Counter == counter {
			return true
		}
	}
	return false
}

func TestShoppingList(t *testing.T) {
	// Create two ShoppingLists
	list1 := NewShoppingList()
	list2 := NewShoppingList()

	// Add or update items in the lists
	list1.AddOrUpdateItem("apple", 3)
	list2.AddOrUpdateItem("banana", 2)

	// Test GetItems function
	items1 := list1.GetItems()
	items2 := list2.GetItems()
	if len(items1) != 1 || items1[0] != "apple" {
		t.Errorf("Expected items1 to be ['apple'], got %v", items1)
	}
	if len(items2) != 1 || items2[0] != "banana" {
		t.Errorf("Expected items2 to be ['banana'], got %v", items2)
	}

	// Test RemoveItem function
	list1.RemoveItem("apple")
	items1AfterRemove := list1.GetItems()
	if len(items1AfterRemove) != 0 {
		t.Errorf("Expected items1AfterRemove to be empty, got %v", items1AfterRemove)
	}

	// Test Merge function
	list1.Merge(list2)
	itemsMerged := list1.GetItems()
	if len(itemsMerged) != 1 || itemsMerged[0] != "banana" {
		t.Errorf("Expected merged items to be ['banana'], got %v", itemsMerged)
	}

	// Test GetItemQuantity function
	quantity, _ := list1.GetItemQuantity("banana")
	if quantity != 2 {
		t.Errorf("Expected quantity to be 2, got %d", quantity)
	}
}

func TestRmvIExistingItem(t *testing.T) {
	awset := NewAWSet()
	nodeID := uuid.New().String()
	itemName := "apple"

	awset.AddI(itemName, nodeID)
	awset.RmvI(itemName)

	if awset.Contains(itemName) {
		t.Errorf("Item %s should be removed, but it exists in the set", itemName)
	}
}

func TestRmvINonExistentItem(t *testing.T) {
	awset := NewAWSet()
	nodeID := uuid.New().String()
	itemName := "apple"
	nonExistentItem := "banana"

	awset.AddI(itemName, nodeID)
	awset.RmvI(nonExistentItem)

	if !awset.Contains(itemName) {
		t.Errorf("Original item %s should still exist, but it's not in the set", itemName)
	}
}

func TestRmvIContextUnchanged(t *testing.T) {
	awset := NewAWSet()
	nodeID := uuid.New().String()
	itemName := "apple"

	awset.AddI(itemName, nodeID)
	ContextBeforeRemoval := awset.Context
	awset.RmvI(itemName)

	if !isEqualContext(awset.Context, ContextBeforeRemoval) {
		t.Error("Context should remain unchanged after item removal")
	}
}

func isEqualContext(Context1, Context2 []ContextItem) bool {
	// Implement your logic to compare Context items
	// This will depend on your specific implementation details
	return true
}

func TestFilterFunction(t *testing.T) {
	nodeID1 := uuid.New().String()
	nodeID2 := uuid.New().String()
	awset1 := NewAWSet()
	awset2 := NewAWSet()

	awset1.State = append(awset1.State, item{"apple", nodeID1, 1})
	awset1.Context = append(awset1.Context, ContextItem{nodeID1, 2})

	awset2.State = append(awset2.State, item{"banana", nodeID2, 1})
	awset2.Context = append(awset2.Context, ContextItem{nodeID2, 2})

	expectedState := make(map[item]bool)
	expectedState[item{"apple", nodeID1, 1}] = true

	filteredState := awset1.Filter(awset2)

	for _, item := range filteredState {
		if !expectedState[item] {
			t.Errorf("Filtered State contains unexpected item: %v", item)
		}
	}
}

func TestMergeWithOverlap(t *testing.T) {
	awset1 := NewAWSet()
	awset2 := NewAWSet()

	nodeID1 := uuid.New().String()
	nodeID2 := uuid.New().String()
	counter1 := uint32(1)
	counter2 := uint32(2)

	awset1.State = append(awset1.State, item{"apple", nodeID1, counter1})
	awset2.State = append(awset2.State, item{"apple", nodeID2, counter2})

	awset1.Merge(awset2)

	if len(awset1.State) != 2 {
		t.Error("Merging should result in a set that contains both items")
	}
}

func TestMergeWithDistinctItems(t *testing.T) {
	awset1 := NewAWSet()
	awset2 := NewAWSet()

	nodeID1 := uuid.New().String()
	nodeID2 := uuid.New().String()
	counter1 := uint32(1)
	counter2 := uint32(2)

	awset1.State = append(awset1.State, item{"apple", nodeID1, counter1})
	awset2.State = append(awset2.State, item{"banana", nodeID2, counter2})

	awset1.Merge(awset2)

	if len(awset1.State) != 2 {
		t.Error("Merging should result in a set that contains both items")
	}
}

func TestMergeWithUniqueItems(t *testing.T) {
	awset1 := NewAWSet()
	awset2 := NewAWSet()

	nodeID := uuid.New().String()
	counter := uint32(1)

	awset1.State = append(awset1.State, item{"apple", nodeID, counter})

	awset1.Merge(awset2)

	if len(awset1.State) != 1 {
		t.Error("Merging with an empty set should not change the first set")
	}

	if !awset1.Contains("apple") {
		t.Error("Merging with an empty set should not change the first set")
	}
}

func TestMergeWithEmptySets(t *testing.T) {
	awset1 := NewAWSet()
	awset2 := NewAWSet()

	awset1.Merge(awset2)

	if len(awset1.State) != 0 {
		t.Error("Merging two empty sets should result in an empty set")
	}
}

func TestElements(t *testing.T) {
	awset := NewAWSet()
	nodeID := uuid.New().String()

	awset.AddI("Apple", nodeID)
	awset.AddI("Banana", nodeID)
	awset.AddI("Apple", nodeID)

	elements := awset.Elements()

	if len(elements) != 2 {
		t.Errorf("Should contain 2 unique items, got %d", len(elements))
	}

	if !containsString(elements, "Apple") || !containsString(elements, "Banana") {
		t.Error("Elements do not contain expected items")
	}
}

func TestContains(t *testing.T) {
	awset := NewAWSet()
	nodeID := uuid.New().String()

	awset.AddI("Apple", nodeID)

	if !awset.Contains("Apple") {
		t.Error("Set should contain 'Apple'")
	}

	if awset.Contains("Banana") {
		t.Error("Set should not contain 'Banana'")
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func TestBoundedPNCounterAssociativity(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		a := NewBoundedPNCounter()
		b := NewBoundedPNCounter()
		c := NewBoundedPNCounter()

		// Populate counters with random operations
		// (You may need to adapt this based on your actual implementation)
		// Example:
		a.Increment("node1", 1)
		b.Decrement("node2", 2)
		c.Increment("node3", 3)

		abC := a.Merge(b).Merge(c)
		aBC := a.Merge(b.Merge(c))

		assert.Equal(t, abC.PositiveCount, aBC.PositiveCount)
		assert.Equal(t, abC.NegativeCount, aBC.NegativeCount)
	}
}

func TestBoundedPNCounterCommutativity(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		a := NewBoundedPNCounter()
		b := NewBoundedPNCounter()

		// Populate counters with random operations
		// (You may need to adapt this based on your actual implementation)
		// Example:
		a.Increment("node1", 1)
		b.Decrement("node2", 2)

		ab := a.Merge(b)
		ba := b.Merge(a)

		assert.Equal(t, ab.PositiveCount, ba.PositiveCount)
		assert.Equal(t, ab.NegativeCount, ba.NegativeCount)
	}
}

func TestBoundedPNCounterIdempotency(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		a := NewBoundedPNCounter()

		// Populate counters with random operations
		// (You may need to adapt this based on your actual implementation)
		// Example:
		a.Increment("node1", 1)
		a.Decrement("node2", 2)

		aa := a.Merge(a)

		assert.Equal(t, a.PositiveCount, aa.PositiveCount)
		assert.Equal(t, a.NegativeCount, aa.NegativeCount)
	}
}

func TestBoundedPNCounterCompare(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		a := NewBoundedPNCounter()
		b := a

		// Add random amounts to both positive and negative counts of `b`
		// (You may need to adapt this based on your actual implementation)
		// Example:
		for nodeID := range a.PositiveCount {
			additionalAmount1 := uint32(1)
			additionalAmount2 := uint32(2)
			b.Increment(nodeID, additionalAmount1)
			b.Decrement(nodeID, additionalAmount2)
		}

		// Invariant: `a` in this Context will be always less or equal to `b` -> so it's comparable
		assert.True(t, a.Compare(b))
	}
}

func TestBoundedPNCounterOverflow(t *testing.T) {
	t.Parallel()

	nodeID := uuid.New()

	for i := 0; i < numberOfProperties; i++ {
		counter := NewBoundedPNCounter()

		// Increment with u32::MAX
		counter.Increment(nodeID.String(), uint32(^uint32(0)))

		// Increment with a random amount
		amount := uint32(rand.Intn(100) + 1)
		counter.Increment(nodeID.String(), amount)

		// Check if value is either u32::MAX or wrapped around
		require.True(t, counter.PositiveCount[nodeID.String()] == uint32(^uint32(0)) ||
			counter.PositiveCount[nodeID.String()] < uint32(^uint32(0)))
	}
}

func TestAWSetAddRemoveProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(100) + 1)
		nodeID := generateRandomUUID()

		awset := NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])
		str := u1.String()

		awset.AddI(item, str)
		if !awset.Contains(item) {
			t.Errorf("Test failed for item: %s, nodeID: %s", item, str)
		}

		awset.RmvI(item)
		if awset.Contains(item) {
			t.Errorf("Test failed for item: %s, nodeID: %s", item, str)
		}
	}
}

func TestAWSetAssociativityAfterAddProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(10))
		nodeID := generateRandomUUID()

		a, b, c := NewAWSet(), NewAWSet(), NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])
		u2, _ := uuid.FromBytes(nodeID[:])
		u3, _ := uuid.FromBytes(nodeID[:])

		a.AddI(item, u1.String())
		b.AddI(item, u2.String())
		c.AddI(item, u3.String())

		abC := a.Clone()
		aBC := a.Clone()
		bClone := b.Clone()

		abC.Merge(b)
		abC.Merge(c)

		bClone.Merge(c)
		aBC.Merge(bClone)

		if !equalStateAndContext(abC, aBC) {
			t.Errorf("Test failed for item: %s, nodeID: %s", item, u1.String())
		}
	}
}

func TestAWSetCommutativityAfterAddProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(100) + 1)
		nodeID := generateRandomUUID()

		a, b := NewAWSet(), NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])
		u2, _ := uuid.FromBytes(nodeID[:])

		a.AddI(item, u1.String())
		b.AddI(item, u2.String())

		ab := a.Clone()
		ba := b.Clone()

		ab.Merge(b)
		ba.Merge(a)

		if !equalStateAndContext(ab, ba) {
			t.Errorf("Test failed for item: %s, nodeID1: %s, nodeID2: %s", item, u1.String(), u2.String())
		}
	}
}

func TestAWSetIdempotenceAfterAddProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(10))
		nodeID := generateRandomUUID()

		a := NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])

		a.AddI(item, u1.String())

		aa := a.Clone()
		aa.Merge(a)

		if !equalStateAndContext(a, aa) {
			t.Errorf("Test failed for item: %s, nodeID: %s", item, u1.String())
		}
	}
}

func TestAWSetAssociativityAfterAddRemoveProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(10))
		nodeID := generateRandomUUID()

		a, b, c := NewAWSet(), NewAWSet(), NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])
		u2, _ := uuid.FromBytes(nodeID[:])
		u3, _ := uuid.FromBytes(nodeID[:])

		a.AddI(item, u1.String())
		b.AddI(item, u2.String())
		c.AddI(item, u3.String())

		a.RmvI(item)

		abC := a.Clone()
		aBC := a.Clone()
		bClone := b.Clone()

		abC.Merge(b)
		abC.Merge(c)

		bClone.Merge(c)
		aBC.Merge(bClone)

		if !equalStateAndContext(abC, aBC) {
			t.Errorf("Test failed for item: %s, nodeID: %s", item, u1.String())
		}
	}
}

func TestAWSetCommutativityAfterAddRemoveProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(100) + 1)
		nodeID := generateRandomUUID()

		a, b := NewAWSet(), NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])
		u2, _ := uuid.FromBytes(nodeID[:])

		a.AddI(item, u1.String())
		b.AddI(item, u2.String())

		b.RmvI(item)

		ab := a.Clone()
		ba := b.Clone()

		ab.Merge(b)
		ba.Merge(a)

		if !equalStateAndContext(ab, ba) {
			t.Errorf("Test failed for item: %s, nodeID1: %s, nodeID2: %s", item, u1.String(), u2.String())
		}
	}
}

func TestAWSetIdempotenceAfterAddRemoveProperty(t *testing.T) {
	t.Parallel()

	for i := 0; i < numberOfProperties; i++ {
		// Generate random inputs
		item := generateRandomString(rand.Intn(100) + 1)
		nodeID := generateRandomUUID()

		a := NewAWSet()
		u1, _ := uuid.FromBytes(nodeID[:])

		a.AddI(item, u1.String())
		a.RmvI(item)

		aa := a.Clone()
		aa.Merge(a)

		if !equalStateAndContext(a, aa) {
			t.Errorf("Test failed for item: %s, nodeID: %s", item, u1.String())
		}
	}
}

func TestAWSetAssociativity(t *testing.T) {
	t.Parallel()

	// Run the test multiple times
	for i := 0; i < numberOfProperties; i++ {
		// Generate random AWSets
		a := generateRandomAWSet()
		abC := a.Clone()
		aBc := a.Clone()

		b := generateRandomAWSet()
		bc := b.Clone()

		c := generateRandomAWSet()

		// Perform the operations
		abC.Merge(b)
		abC.Merge(c)

		bc.Merge(c)
		aBc.Merge(bc)

		// Check if the state and context are equal
		if !equalStateAndContext(abC, aBc) {
			t.Errorf("Test failed on iteration %d. State and context are not equal.", i)
		}
	}
}

func TestAWSetAssociativitySmall(t *testing.T) {
	t.Parallel()
	for i := 0; i < numberOfProperties; i++ {
		a := NewAWSet()
		b := NewAWSet()
		c := NewAWSet()

		a.AddI("a", "1")
		b.AddI("b", "2")
		c.AddI("c", "3")

		abC := a.Clone()
		aBc := a.Clone()

		abC.Merge(b)
		abC.Merge(c)

		bc := b.Clone()
		bc.Merge(c)

		aBc.Merge(bc)

		if !equalStateAndContext(abC, aBc) {
			t.Errorf("Test failed on iteration %d. State and context are not equal.", i)
		}

	}
}

func TestAWSetCommutativity(t *testing.T) {
	t.Parallel()

	// Run the test multiple times
	for i := 0; i < numberOfProperties; i++ {
		// Generate random AWSets
		a := generateRandomAWSet()
		b := generateRandomAWSet()

		// Perform the operations
		ab := a.Clone()
		ab.Merge(b)

		ba := b.Clone()
		ba.Merge(a)

		// Check if the state and context are equal
		if !equalStateAndContext(ab, ba) {
			t.Errorf("Test failed on iteration %d. State and context are not equal.", i)
		}
	}
}

func TestAwSetIdempotence(t *testing.T) {
	t.Parallel()

	// Run the test multiple times
	for i := 0; i < numberOfProperties; i++ {
		// Generate random AWSets
		a := generateRandomAWSet()

		// Perform the operations
		aa := a.Clone()
		aa.Merge(a)

		// Check if the state and context are equal
		if !equalStateAndContext(a, aa) {
			t.Errorf("Test failed on iteration %d. State and context are not equal.", i)
		}
	}
}

func TestConvergence(t *testing.T) {
	t.Parallel()

	// Run the test multiple times
	for i := 0; i < numberOfProperties; i++ {
		// Generate random AWSets
		a := generateRandomAWSet()
		b := generateRandomAWSet()
		c := generateRandomAWSet()

		ab := a.Clone()
		ac := a.Clone()

		ab.Merge(b)
		ac.Merge(c)

		bc := b.Clone()
		bc.Merge(c)

		ab.Merge(bc)
		ac.Merge(bc)

		// Check if the state and context are equal
		if !equalStateAndContext(ab, ac) {
			t.Errorf("Test failed on iteration %d. State and context are not equal.", i)
		}
	}
}

func TestElementAdditionRemoval(t *testing.T) {
	t.Parallel()
	// Run the test multiple times
	for i := 0; i < numberOfProperties; i++ {
		// Generate random AWSets
		a := generateRandomAWSet()

		//add one random element
		item := generateRandomString(rand.Intn(100) + 1)
		nodeID := generateRandomUUID()
		u1, _ := uuid.FromBytes(nodeID[:])
		a.AddI(item, u1.String())

		//check if element is in set
		if !a.Contains(item) {
			t.Errorf("Test failed on iteration %d. Element %s is not in set.", i, item)
		}

		//remove element
		a.RmvI(item)

		//check if element is in set
		if a.Contains(item) {
			t.Errorf("Test failed on iteration %d. Element %s is in set.", i, item)
		}

	}
}

func TestShoppingListAssociativity(t *testing.T) {
	t.Parallel()
	for i := 0; i < numberOfProperties; i++ {
		a := generateRandomShoppingList()
		b := generateRandomShoppingList()
		c := generateRandomShoppingList()

		abC := a.Clone()
		aBc := a.Clone()

		abC.Merge(b)
		abC.Merge(c)

		bc := b.Clone()
		bc.Merge(c)

		aBc.Merge(bc)
		if !equalShoppingList(abC, aBc) {
			t.Errorf("Test failed on iteration %d. Shopping lists are not equal.", i)
		}
	}
}

func TestShoppingListCommutativity(t *testing.T) {
	t.Parallel()
	for i := 0; i < numberOfProperties; i++ {
		a := generateRandomShoppingList()
		b := generateRandomShoppingList()

		ab := a.Clone()
		ab.Merge(b)

		ba := b.Clone()
		ba.Merge(a)

		if !equalShoppingList(ab, ba) {
			t.Errorf("Test failed on iteration %d. Shopping lists are not equal.", i)
		}
	}
}

func TestShoppingListIdempotence(t *testing.T) {
	t.Parallel()
	for i := 0; i < numberOfProperties; i++ {
		a := generateRandomShoppingList()

		aa := a.Clone()
		aa.Merge(a)

		if !equalShoppingList(a, aa) {
			t.Errorf("Test failed on iteration %d. Shopping lists are not equal.", i)
		}
	}
}

func TestShoppingListAddUpdateRemove(t *testing.T) {
	t.Parallel()
	for i := 0; i < numberOfProperties; i++ {
		a := generateRandomShoppingList()

		//add one random element
		item := generateRandomString(rand.Intn(100) + 1)
		quantity := rand.Intn(100) + 1
		a.AddOrUpdateItem(item, quantity)

		aOriginal := a.Clone()

		//check if element is in set
		if q, ok := a.GetItemQuantity(item); !ok || q != int32(quantity) {
			t.Errorf("Test failed on iteration %d. Element %s is not in set.", i, item)
		}

		//remove element
		a.RemoveItem(item)

		//check if element is in set
		if q, ok := a.GetItemQuantity(item); ok || q != int32(0) {
			t.Errorf("Test failed on iteration %d. Element %s is in set.", i, item)
		}

		//check if original list is unchanged
		if !(!equalItemsMap(a.Items, aOriginal.Items) &&
			equalContextItems(a.AwSet.Context, aOriginal.AwSet.Context) &&
			!equalItems(a.AwSet.State, aOriginal.AwSet.State)) {

			t.Errorf("Test failed on iteration %d. Shopping lists are not equal.", i)
		}

	}
}

// Utility function to compare State and Context
func equalStateAndContext(a, b *AWSet) bool {
	return equalItems(a.State, b.State) && equalContextItems(a.Context, b.Context)
}

func equalItems(a, b []item) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to store counts of items
	countA := make(map[item]int)
	countB := make(map[item]int)

	// Count items in slice A
	for _, item := range a {
		countA[item]++
	}

	// Count items in slice B
	for _, item := range b {
		countB[item]++
	}

	// Compare counts
	for item, count := range countA {
		if countB[item] != count {
			return false
		}
	}

	return true
}

func equalContextItems(a, b []ContextItem) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to store counters for each node ID
	counterMapA := make(map[string]uint32)
	counterMapB := make(map[string]uint32)

	// Populate counter maps for slice a
	for _, ctxItem := range a {
		counterMapA[ctxItem.NodeID] = ctxItem.Counter
	}

	// Populate counter maps for slice b
	for _, ctxItem := range b {
		counterMapB[ctxItem.NodeID] = ctxItem.Counter
	}

	// Compare counters for each node ID
	for nodeID, counterA := range counterMapA {
		counterB, ok := counterMapB[nodeID]
		if !ok || counterA != counterB {
			return false
		}
	}

	return true
}

// equalItemsMap checks if two maps of items are equal without relying on order.
func equalItemsMap(a, b map[string]*BoundedPNCounter) bool {
	if len(a) != len(b) {
		return false
	}

	for itemName, counterA := range a {
		counterB, ok := b[itemName]
		if !ok || !counterA.Compare(counterB) {
			return false
		}
	}

	for itemName, counterB := range b {
		counterA, ok := a[itemName]
		if !ok || !counterB.Compare(counterA) {
			return false
		}
	}
	return true
}

// equalShoppingList checks if two ShoppingLists are equal in terms of items, state, and context.
func equalShoppingList(a, b *ShoppingList) bool {
	return equalItemsMap(a.Items, b.Items) &&
		equalItems(a.AwSet.State, b.AwSet.State) &&
		equalContextItems(a.AwSet.Context, b.AwSet.Context)
}

// Utility function to generate a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

// Utility function to generate a random UUID
func generateRandomUUID() [16]byte {
	return uuid.New()
}

func generateRandomAWSet() *AWSet {
	awSet := NewAWSet()

	// Generate a random number of items
	numItems := rand.Intn(100) + 1 // You can adjust the range as needed

	for i := 0; i < numItems; i++ {
		// Generate a random name
		itemName := generateRandomString(30)

		// Generate a random node ID
		nodeID := generateRandomUUID()
		u1, _ := uuid.FromBytes(nodeID[:])
		// Add the item to the AWSet using the existing AddI method
		numCounter := rand.Intn(30) + 1

		for j := 0; j < numCounter; j++ {
			awSet.AddI(itemName, u1.String())
		}
	}

	return awSet
}

// generateRandomShoppingList generates a random ShoppingList.
func generateRandomShoppingList() *ShoppingList {
	list := NewShoppingList()
	// Generate a random number of items
	numItems := rand.Intn(100) + 1 // You can adjust the range as needed

	for i := 0; i < numItems; i++ {
		// Generate a random name
		itemName := generateRandomString(30)

		// Generate a random quantity change
		quantityChange := rand.Intn(100) + 1

		list.AddOrUpdateItem(itemName, quantityChange)

		quantityChange = rand.Intn(30) * -1

		list.AddOrUpdateItem(itemName, quantityChange)
	}

	return list
}
