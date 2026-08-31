package main

import (
	"database/sql"
	"testing"
)

func mkNode(id int) *RatingTreeNode {
	return &RatingTreeNode{Review: MovieReview{ReviewID: id}}
}

func link(parent, left, right *RatingTreeNode) {
	parent.Left = left
	parent.Right = right
	if left != nil {
		left.Parent = parent
	}
	if right != nil {
		right.Parent = parent
	}
}

func idOf(n *RatingTreeNode) int {
	if n == nil {
		return -1
	}
	return n.Review.ReviewID
}

// Tree:
//              10
//           /      \
//          5        15
//         / \       /  \
//        3   8     12   20
//           /
//          7
//         /
//        6
func buildSample() (root *RatingTreeNode, nodes map[int]*RatingTreeNode) {
	nodes = map[int]*RatingTreeNode{}
	for _, id := range []int{10, 5, 15, 3, 8, 12, 20, 7, 6} {
		nodes[id] = mkNode(id)
	}
	root = nodes[10]
	link(nodes[10], nodes[5], nodes[15])
	link(nodes[5], nodes[3], nodes[8])
	link(nodes[15], nodes[12], nodes[20])
	link(nodes[8], nodes[7], nil)
	link(nodes[7], nodes[6], nil)
	return root, nodes
}

func TestDeleteLeaf(t *testing.T) {
	root, nodes := buildSample()
	_ = root
	changed := DeleteNode(nodes[3])
	if nodes[5].Left != nil {
		t.Fatalf("expected 5.Left nil, got %v", idOf(nodes[5].Left))
	}
	found := false
	for _, n := range changed {
		if n == nodes[5] {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected parent 5 in changed set")
	}
}

func TestDeleteOneChild(t *testing.T) {
	root, nodes := buildSample()
	_ = root
	// delete 8 (has one child: 7)
	DeleteNode(nodes[8])
	if nodes[5].Right != nodes[7] {
		t.Fatalf("expected 5.Right == 7, got %v", idOf(nodes[5].Right))
	}
	if nodes[7].Parent != nodes[5] {
		t.Fatalf("expected 7.Parent == 5, got %v", idOf(nodes[7].Parent))
	}
	if nodes[7].Left != nodes[6] {
		t.Fatalf("expected 7.Left still == 6, got %v", idOf(nodes[7].Left))
	}
}

func TestDeleteTwoChildrenDeepSuccessor(t *testing.T) {
	root, nodes := buildSample()
	// delete root 10; successor should be leftmost of right subtree (15) -> 12
	changed := DeleteNode(nodes[10])
	newRoot := findNode(root, -1) // root itself was deleted; find via any surviving node's ancestor chain
	_ = newRoot

	successor := nodes[12]
	if successor.Parent != nil {
		t.Fatalf("expected successor 12 to become root (nil parent), got parent %v", idOf(successor.Parent))
	}
	if successor.Left != nodes[5] {
		t.Fatalf("expected 12.Left == 5, got %v", idOf(successor.Left))
	}
	if nodes[5].Parent != successor {
		t.Fatalf("expected 5.Parent == 12, got %v", idOf(nodes[5].Parent))
	}
	if successor.Right != nodes[15] {
		t.Fatalf("expected 12.Right == 15, got %v", idOf(successor.Right))
	}
	if nodes[15].Parent != successor {
		t.Fatalf("expected 15.Parent == 12, got %v", idOf(nodes[15].Parent))
	}
	if nodes[15].Left != nil {
		t.Fatalf("expected 15.Left nil (12 detached), got %v", idOf(nodes[15].Left))
	}
	if nodes[15].Right != nodes[20] {
		t.Fatalf("expected 15.Right still == 20, got %v", idOf(nodes[15].Right))
	}

	// sanity: changed set should include successor and its new parent(none), plus 15 and 5
	names := map[int]bool{}
	for _, n := range changed {
		names[n.Review.ReviewID] = true
	}
	for _, want := range []int{12, 15, 5} {
		if !names[want] {
			t.Fatalf("expected %d in changed set, got %v", want, names)
		}
	}
}

func TestDeleteTwoChildrenDirectRightSuccessor(t *testing.T) {
	nodes := map[int]*RatingTreeNode{}
	for _, id := range []int{10, 5, 15, 20} {
		nodes[id] = mkNode(id)
	}
	link(nodes[10], nodes[5], nodes[15])
	link(nodes[15], nil, nodes[20]) // 15 has no left child -> is its own successor when deleting 10

	DeleteNode(nodes[10])

	if nodes[15].Parent != nil {
		t.Fatalf("expected 15 to become root, got parent %v", idOf(nodes[15].Parent))
	}
	if nodes[15].Left != nodes[5] {
		t.Fatalf("expected 15.Left == 5, got %v", idOf(nodes[15].Left))
	}
	if nodes[5].Parent != nodes[15] {
		t.Fatalf("expected 5.Parent == 15, got %v", idOf(nodes[5].Parent))
	}
	if nodes[15].Right != nodes[20] {
		t.Fatalf("expected 15.Right to remain 20, got %v", idOf(nodes[15].Right))
	}
	if nodes[20].Parent != nodes[15] {
		t.Fatalf("expected 20.Parent still == 15, got %v", idOf(nodes[20].Parent))
	}
}

func TestPtrID(t *testing.T) {
	n := mkNode(42)
	got := ptrID(n)
	want := sql.NullInt64{Int64: 42, Valid: true}
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if ptrID(nil).Valid {
		t.Fatalf("expected ptrID(nil) to be invalid")
	}
}
