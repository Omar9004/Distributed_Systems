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

func (cr *ChordRing) Notify(nPrime string) bool {
	if cr.Predecessor != "" {
		requestInfo := FindSucRequest{}
		requestInfo.InfoType = GetID
		getPredRep := cr.CallFS(cr.Predecessor, "ChordRing.GetNodeInfo", &requestInfo) //Call the predecessor node to get its ID(Identifier)
		predID := getPredRep.Identifier                                                //Predecessor ID

		getNodeID := cr.CallFS(nPrime, "ChordRing.GetNodeInfo", &requestInfo)
		newNodeID := getNodeID.Identifier

		fmt.Printf("PredID:%v\n", predID)
		fmt.Printf("newNodeID:%v\n", newNodeID)
		fmt.Printf("cr.Identifier:%v\n", cr.Identifier)
		isBetween := between(predID, newNodeID, cr.Identifier, false)
		if isBetween {
			cr.Predecessor = nPrime
			return true
		} else {
			return false
		}

	} else {
		fmt.Printf("cr.Predecessor before assigning:%v\n", cr.Predecessor)

		cr.Predecessor = nPrime

		return true
	}
}

func (cr *ChordRing) NotifyRPC(args *NotifyArgs, replay *NotifyReply) error {
	if cr.Notify(args.NewIPAddress) {
		replay.isComplete = true
	} else {
		replay.isComplete = false

	}
	return nil
}
func (cr *ChordRing) JoinChord(joinNodeAdd string, args *FindSucRequest, replay *FindSucReplay) error {
	//extractIP := strings.Split(joinNodeAdd, ":")[0]
	//joinId := IdentifierGen(extractIP)
	getReq := FindSucRequest{-1,
		cr.Identifier,
		cr.FullAddress,
	}
	fmt.Println("joinNodeAdd", joinNodeAdd)
	findSucReply := cr.CallFS(joinNodeAdd, "ChordRing.FindSuccessor", &getReq)

	notifyReq := NotifyArgs{}
	cr.Successors[0] = findSucReply.SuccAddress
	fmt.Printf("Found the successor: %s\n", replay.SuccAddress)
	notifyReq.NewIPAddress = cr.Successors[0]

	cr.CallNotify(cr.Successors[0], "ChordRing.NotifyRPC", &notifyReq)
	getPre := FindSucRequest{}
	getPre.InfoType = GetPre
	infoPre := cr.CallFS(cr.Successors[0], "ChordRing.GetNodeInfo", &getPre)
	fmt.Printf("The predecessor node: %v\n", infoPre.Predecessor)
	fmt.Printf("The current node: %s\n", cr.FullAddress)
	fmt.Printf("The successor node: %s\n", cr.Successors[0])

	return nil
}
