package lab3_Chord

import (
	"crypto/sha1"
	"math/big"
)

type Key string

type NodeAddress string

const m = 6

type Node struct {
	Identifier  *big.Int
	FingerTable []*Node
	Predecessor *Node
	Successors  *Node

	Bucket map[Key]string
}

// CreateRing initializes the Chord Ring
func (n *Node) CreateRing() {
	n.Predecessor = nil
	n.Successors = n
	n.FingerTable = make([]*Node, m)
}

func (n *Node) Join(current *Node) {
	n.Successors = current
}

func (n *Node) FindSuccessor(key *big.Int) *Node {

}
func (n *Node) NewNode(identifier big.Int) *Node {

}
