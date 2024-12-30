package main

import (
	"fmt"
	"math/big"
)

const m = 6

type finger struct {
	Identifier *big.Int
	IPAddress  string
}
type ChordRing struct {
	//LocalNode      Node    //Local node that acts as entry point for the chord operations
	//Nodes          []*Node //Maintains a list of nodes exist in the ring
	HashBits       int //Defines the size of the hashing space
	StabilizeTime  int
	FixFingersTime int
	CheckPredTime  int
	SuccessorNum   int

	//Node's instances
	Identifier  *big.Int
	FingerTable []finger
	Predecessor string
	Successors  []string
	IPAddress   string
	FullAddress string //IP address and Channel Port
	Bucket      map[Key]string
}

func (cr *ChordRing) NewChordRing() {
	//cr.Nodes = make([]*Node, int(math.Pow(2, m)))
	//fmt.Printf("Chord Ring length: %d\n", len(cr.Nodes))

	//cr.LocalNode = startNode
	//id := startNode.Identifier.Int64()
	//cr.Nodes[id] = &startNode
	cr.Predecessor = ""
	cr.Successors[0] = cr.FullAddress
	//cr.Nodes[id].Predecessor = startNode.IPAddress

}

func (cr *ChordRing) lookup(key *big.Int, stratingNode string) {

}

func (cr *ChordRing) StoreFile(fileName string) error {
	return nil
}

//func (cr *ChordRing) AddIdentifier(node ) {
//	cr.Nodes = append(cr.Nodes, node)
//}

func (cr *ChordRing) Notify(args *FindSucRequest, replay *FindSucReplay) {

}
func (cr *ChordRing) JoinChord(joinNodeAdd string, args *FindSucRequest, replay *FindSucReplay) error {
	//extractIP := strings.Split(joinNodeAdd, ":")[0]
	//joinId := IdentifierGen(extractIP)
	findSuc := FindSucReplay{}
	getReq := FindSucRequest{}
	getReq.IPAddress = cr.IPAddress
	getReq.Identifier = cr.Identifier
	fmt.Println("joinNodeAdd", joinNodeAdd)
	if !cr.call(joinNodeAdd, "ChordRing.FindSuccessor", &getReq, &findSuc) {
		return fmt.Errorf("failed to find successor for node %s", joinNodeAdd)
	}
	newReq := FindSucRequest{}
	cr.Successors[0] = replay.SuccAddress
	newReq.IPAddress = cr.Successors[0]

	if !cr.call(cr.Successors[0], "ChordRing.Notify", newReq, replay) {
		return fmt.Errorf("the notification call is unsuccessful %s", cr.Successors[0])
	}

	return nil
}
