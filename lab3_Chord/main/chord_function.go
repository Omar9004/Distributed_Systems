package main

import (
	"fmt"
	"math"
	"math/big"
)

const m = 6

type ChordRing struct {
	LocalNode      *Node   //Local node that acts as entry point for the chord operations
	Nodes          []*Node //Maintains a list of nodes exist in the ring
	HashBits       int     //Defines the size of the hashing space
	StabilizeTime  int
	FixFingersTime int
	CheckPredTime  int
	SuccessorNum   int
}

func NewChordRing(startNode *Node) *ChordRing {
	cr := ChordRing{}
	cr.Nodes = make([]*Node, int(math.Pow(2, m)))
	fmt.Printf("Chord Ring length: %d\n", len(cr.Nodes))
	cr.Nodes[0] = startNode
	return &cr
}

func (cr *ChordRing) JoinChord(localNode *Node) {
	cr.Nodes = append(cr.Nodes, localNode)
}
func (cr *ChordRing) lookup(key *big.Int, stratingNode string) *Node {
	return nil

}

func (cr *ChordRing) StoreFile(fileName string, node *Node) error {
	return nil
}

func (cr *ChordRing) AddIdentifier(node *Node) {
	cr.Nodes = append(cr.Nodes, node)
}
func (n *Node) FindSuccessor(key *big.Int) error {

	return nil
}
