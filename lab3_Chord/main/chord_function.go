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
	getReq := FindSucRequest{
		cr.Identifier,
		cr.FullAddress,
	}
	fmt.Println("joinNodeAdd", joinNodeAdd)
	findSucReply := cr.MakeCall(joinNodeAdd, "ChordRing.FindSuccessor", &getReq)

	notifyReq := FindSucRequest{}
	cr.Successors[0] = findSucReply.SuccAddress
	fmt.Printf("Found the successor: %s\n", replay.SuccAddress)
	notifyReq.IPAddress = cr.Successors[0]

	notifyReplay := cr.MakeCall(cr.Successors[0], "ChordRing.Notify", &notifyReq)
	fmt.Println(notifyReplay)

	return nil
}
