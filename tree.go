package main

import "fmt"

func BuildTree(reviews []*MovieReview) *RatingTreeNode {
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
