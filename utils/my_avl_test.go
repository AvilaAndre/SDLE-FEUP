package utils

import (
	"testing"
)

func TestAVLTree(t *testing.T) {
	tree := &AVLTree{}

	tree.Add("h", "1")
	tree.Add("b", "2")
	tree.Add("c", "3")
	tree.Add("d", "4")
	tree.Add("e", "5")
	tree.Add("f", "6")
	tree.Add("k", "7")
	tree.Add("a", "8")
	tree.Add("i", "9")
	tree.Add("j", "10")

	// Should get the next key if the target does not exist
	if tree.Search("g").key != "h" {
		t.Fail()
	}

	if tree.Search("h").key != "h" {
		t.Fail()
	}

	// Should get the first key if last one is not the target
	if tree.Search("x").key != "a" {
		t.Fail()
	}

	// Should get the next key after the specified key
	if tree.Next("a").key != "b" {
		t.Log("next key after a should be b, not", tree.Next("a").key)
		t.Fail()
	}

	if tree.Next("x").key != "a" {
		t.Log("next key after x should be a, not", tree.Next("x").key)
		t.Fail()
	}

	if tree.Next("c").key != "d" {
		t.Log("next key after c should be d, not", tree.Next("c").key)
		t.Fail()
	}

}
