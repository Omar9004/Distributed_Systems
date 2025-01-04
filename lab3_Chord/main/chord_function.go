package main

import (
	"fmt"
	"math/big"
	"net/rpc"
	"time"
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
	StabilizeTime  time.Duration
	FixFingersTime time.Duration
	CheckPredTime  time.Duration
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

func (cr *ChordRing) Stabilize() error {

	//1. Check on the Successor's predecessor pointer whether it is point back to the current node or not.
	//By calling the successor's predecessor node
	//*//

	newReq := FindSucRequest{}
	newReq.InfoType = GetPre // Get the Successor's predecessor
	preReplay, err := cr.CallStabilize(cr.Successors[0], "ChordRing.GetNodeInfo", &newReq)
	fmt.Printf("The predeceussor node of the successor: %v\n", preReplay.Predecessor)
	if err == nil && preReplay.Predecessor != "" {
		newReq.InfoType = GetID

		sucReplay, _ := cr.CallStabilize(preReplay.Predecessor, "ChordRing.GetNodeInfo", &newReq)

		sucId := IdentifierGen(cr.Successors[0]) //Extract the successor's ID from its ip address
		sucPred := sucReplay.Identifier          //Successor's predecessor ID
		isBetween := between(cr.Identifier, sucPred, sucId, false)
		fmt.Printf("isBetween the predecessor and the current node's successor: %v\n", isBetween)
		if isBetween {
			cr.Successors[0] = preReplay.Predecessor
			fmt.Printf("New successor: %s, Current Node: %s\n", cr.Successors[0], cr.FullAddress)
		}
	}
	//Notify the successor about its predecessor
	notifyReq := NotifyArgs{}
	notifyReq.NewIPAddress = cr.FullAddress
	cr.CallNotify(cr.Successors[0], "ChordRing.NotifyRPC", &notifyReq)

	return nil
}

func (cr *ChordRing) Check_predecessor() {
	if cr.Predecessor != "" {
		_, err := rpc.Dial("tcp", cr.Predecessor)
		if err != nil {
			fmt.Printf("The Predecessor:%s of the node: %s is nolonger avalaible\n", cr.Predecessor, cr.FullAddress)
			cr.Predecessor = ""
		}
	}
}
