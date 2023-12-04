package crdt_go

import (
	"testing"
)

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
	quantity := list1.GetItemQuantity("banana")
	if quantity != 2 {
		t.Errorf("Expected quantity to be 2, got %d", quantity)
	}
}
