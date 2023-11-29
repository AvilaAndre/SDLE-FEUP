package utils

import (
	"fmt"
)

// An AVL tree in golang was already implemented here (https://github.com/karask/go-avltree), however, I want to make some changes to it

// AVLTree structure. Public methods are Add, Remove, Update, Search, DisplayTreeInOrder.
type AVLTree struct {
	root *AVLNode
}

func (t *AVLTree) Add(key string, value string) {
	t.root = t.root.add(key, value)
}

func (t *AVLTree) Remove(key string) {
	t.root = t.root.remove(key)
}

func (t *AVLTree) Update(oldKey string, newKey string, newValue string) {
	t.root = t.root.remove(oldKey)
	t.root = t.root.add(newKey, newValue)
}

func (t *AVLTree) Search(key string) (node *AVLNode) {
	searchResult := t.root.search(key)
	if searchResult != nil {
		return searchResult
	} else {
		if t.root == nil {
			return nil
		}
		return t.root.findSmallest()
	}
}

func (t *AVLTree) Next(key string) (node *AVLNode) {
	nextResult := t.root.next(key)
	if nextResult != nil {
		return nextResult
	} else {
		return t.root.findSmallest()
	}
}

func (t *AVLTree) DisplayInOrder() {
	t.root.displayNodesInOrder()
	fmt.Println()
}

// AVLNode structure
type AVLNode struct {
	key   string
	Value string

	// height counts nodes (not edges)
	height int
	left   *AVLNode
	right  *AVLNode
}

// Adds a new node
func (n *AVLNode) add(key string, value string) *AVLNode {
	if n == nil {
		return &AVLNode{key, value, 1, nil, nil}
	}

	if key < n.key {
		n.left = n.left.add(key, value)
	} else if key > n.key {
		n.right = n.right.add(key, value)
	} else {
		// if same key exists update value
		n.Value = value
	}
	return n.rebalanceTree()
}

// Removes a node
func (n *AVLNode) remove(key string) *AVLNode {
	if n == nil {
		return nil
	}
	if key < n.key {
		n.left = n.left.remove(key)
	} else if key > n.key {
		n.right = n.right.remove(key)
	} else {
		if n.left != nil && n.right != nil {
			// node to delete found with both children;
			// replace values with smallest node of the right sub-tree
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			n.Value = rightMinNode.Value
			// delete smallest node that we replaced
			n.right = n.right.remove(rightMinNode.key)
		} else if n.left != nil {
			// node only has left child
			n = n.left
		} else if n.right != nil {
			// node only has right child
			n = n.right
		} else {
			// node has no children
			n = nil
			return n
		}

	}
	return n.rebalanceTree()
}

// Searches for a node
func (n *AVLNode) search(key string) *AVLNode {
	if n == nil {
		return nil
	}
	if key < n.key {
		// Get the next key if not found
		searchResult := n.left.search(key)
		if searchResult == nil {
			return n
		} else {
			return searchResult
		}
	} else if key > n.key {
		return n.right.search(key)
	} else {
		return n
	}
}

// Finds the next node
func (n *AVLNode) next(key string) *AVLNode {
	if n == nil {
		// nil signals that it is not this node
		return nil
	}
	if key < n.key {
		// Get the next key if not found
		nextResult := n.left.next(key)
		if nextResult == nil {
			return n
		} else {
			return nextResult
		}
	} else if key > n.key {
		// Get the next key if not found
		return n.right.next(key)
	} else {
		if n.right != nil {
			return n.right.findSmallest()
		} else {
			// nil signals that it is not this node
			return nil
		}
	}
}

// Displays nodes left-depth first (used for debugging)
func (n *AVLNode) displayNodesInOrder() {
	if n.left != nil {
		n.left.displayNodesInOrder()
	}
	fmt.Print(n.key, " ")
	if n.right != nil {
		n.right.displayNodesInOrder()
	}
}

func (n *AVLNode) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *AVLNode) recalculateHeight() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// Checks if node is balanced and rebalance
func (n *AVLNode) rebalanceTree() *AVLNode {
	if n == nil {
		return n
	}
	n.recalculateHeight()

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor == -2 {
		// check if child is left-heavy and rotateRight first
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	} else if balanceFactor == 2 {
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	}
	return n
}

// Rotate nodes left to balance node
func (n *AVLNode) rotateLeft() *AVLNode {
	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Rotate nodes right to balance node
func (n *AVLNode) rotateRight() *AVLNode {
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Finds the smallest child (based on the key) for the current node
func (n *AVLNode) findSmallest() *AVLNode {
	if n.left != nil {
		return n.left.findSmallest()
	} else {
		return n
	}
}
