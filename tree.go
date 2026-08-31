package main

import (
	"database/sql"
	"fmt"
	"log"
)

func BuildTree(reviews []*MovieReview) *RatingTreeNode {
	log.Println("Building Tree...")
	if len(reviews) == 0 {
		return nil
	}

	// create a node for each review, mapped by ReviewID
	nodeMap := make(map[int]*RatingTreeNode)
	for _, review := range reviews {
		nodeMap[review.ReviewID] = &RatingTreeNode{
			Review: *review,
		}
	}

	// wire up the pointers
	var root *RatingTreeNode
	for _, review := range reviews {
		node := nodeMap[review.ReviewID]

		if review.LeftPtr.Valid {
			node.Left = nodeMap[int(review.LeftPtr.Int64)]
		}
		if review.RightPtr.Valid {
			node.Right = nodeMap[int(review.RightPtr.Int64)]
		}
		if review.ParentPtr.Valid {
			node.Parent = nodeMap[int(review.ParentPtr.Int64)]
		} else {
			// no parent means this is the root
			root = node
		}
	}

	log.Println("current tree:")
	PrintTree(root)

	return root
}

// PrintTree writes an ASCII visualization of a rating tree to stdout,
// rotated 90°: right subtree drawn above the node, left subtree below.
func PrintTree(root *RatingTreeNode) {
	if root == nil {
		fmt.Println("(empty tree)")
		return
	}
	printSubtree(root.Right, "", false)
	fmt.Println(nodeLabel(root))
	printSubtree(root.Left, "", true)
}

func printSubtree(node *RatingTreeNode, prefix string, isLeft bool) {
	if node == nil {
		return
	}

	childPrefix := prefix + "    "
	siblingPrefix := prefix + "│   "

	if isLeft {
		printSubtree(node.Right, siblingPrefix, false)
	} else {
		printSubtree(node.Right, childPrefix, false)
	}

	connector := "┌── "
	if isLeft {
		connector = "└── "
	}
	fmt.Println(prefix + connector + nodeLabel(node))

	if isLeft {
		printSubtree(node.Left, childPrefix, true)
	} else {
		printSubtree(node.Left, siblingPrefix, true)
	}
}

func nodeLabel(node *RatingTreeNode) string {
	return fmt.Sprintf("%s (id=%d)", node.Review.MovieName, node.Review.ReviewID)
}

// findNode returns the node with the given review ID within root's subtree, or nil.
func findNode(root *RatingTreeNode, reviewID int) *RatingTreeNode {
	if root == nil {
		return nil
	}
	if root.Review.ReviewID == reviewID {
		return root
	}
	if found := findNode(root.Left, reviewID); found != nil {
		return found
	}
	return findNode(root.Right, reviewID)
}

// leftmost returns the leftmost (smallest) node in the subtree rooted at node.
func leftmost(node *RatingTreeNode) *RatingTreeNode {
	for node.Left != nil {
		node = node.Left
	}
	return node
}

// replaceChild rewires parent's Left/Right pointer (whichever currently points
// at oldChild) to point at newChild instead. No operation if parent is nil.
func replaceChild(parent, oldChild, newChild *RatingTreeNode) {
	if parent == nil {
		return
	}
	if parent.Left == oldChild {
		parent.Left = newChild
	} else if parent.Right == oldChild {
		parent.Right = newChild
	}
}

// ptrID converts a node into the sql.NullInt64 form its review_id takes in a
// left_ptr/right_ptr/parent_ptr column (NULL for a nil node).
func ptrID(node *RatingTreeNode) sql.NullInt64 {
	if node == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(node.Review.ReviewID), Valid: true}
}

// DeleteNode removes target from the tree it belongs to, rewiring Left/Right/Parent
// pointers to preserve BST structure (standard leaf / one-child / two-children-via-
// in-order-successor cases). It returns every node whose pointers changed as a
// result, so the caller can persist them; target itself is not included and must
// be deleted separately.
// ahhh this makes me feel like I'm back at uni...
func DeleteNode(target *RatingTreeNode) (changed []*RatingTreeNode) {
	parent := target.Parent
	var replacement *RatingTreeNode

	switch {
	case target.Left == nil && target.Right == nil: // node has no children
		replacement = nil

	case target.Left == nil || target.Right == nil: // node has one child
		if target.Left != nil {
			replacement = target.Left
		} else {
			replacement = target.Right
		}

	default: // node has 2 children
		// inorder successor:
		successor := leftmost(target.Right)
		successorParent := successor.Parent
		successorChild := successor.Right // successor has no Left child by definition

		// if successor is target.Right, then no left descent needed!
		// so we can skip this "unhooking" stage
		if successorParent != target {
			replaceChild(successorParent, successor, successorChild)
			changed = append(changed, successorParent)
			// if successor (right) child exists, then set parent of child so that is isn't orphaned
			if successorChild != nil {
				successorChild.Parent = successorParent
				changed = append(changed, successorChild)
			}

			successor.Right = target.Right
			target.Right.Parent = successor
		}

		successor.Left = target.Left
		target.Left.Parent = successor
		changed = append(changed, target.Left)

		replacement = successor
	}

	// if there is a replacement, then point its parent val to the targets old parent
	if replacement != nil {
		replacement.Parent = parent
		changed = append(changed, replacement)
	}
	// points parent to replacement node (instead of target)
	replaceChild(parent, target, replacement)
	if parent != nil {
		changed = append(changed, parent)
	}

	return changed
}
