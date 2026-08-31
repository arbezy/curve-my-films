package main

import "fmt"

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
