package main

import (
	"fmt"
	"log"
	"math/big"
	"net/rpc"
	"sync"
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
	mutex          sync.Mutex
	//Node's instances
	Identifier  *big.Int
	FingerTable []finger
	Predecessor string
	Successors  []string
	IPAddress   string
	FullAddress string //IP address and Channel Port
	NodeFolder  string //Referring to the node's folder directory
	Bucket      map[*big.Int]string
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

func (cr *ChordRing) lookup(FileName string, NodeAddress string) (*big.Int, string) {
	key := hashString(FileName)
	key.Mod(key, hashMod)
	lookupReq := FindSucRequest{Identifier: key}

	fmt.Printf("lookupReq.Identifier: %s\n", lookupReq.Identifier)
	//newReplay := CallFS(NodeAddress, "ChordRing.FindSuccessor", &lookupReq)
	newReplay := MakeCall[FindSucRequest, FindSucReplay](NodeAddress, "ChordRing.FindSuccessor", lookupReq)
	return key, newReplay.SuccAddress
}

func (cr *ChordRing) storeFile(key *big.Int, SucIP string, fileName string) error {
	newReq := StoreFileArgs{FileName: fileName, Key: key}
	newRep := MakeCall[StoreFileArgs, StoreFileReply](SucIP, "ChordRing.StoreFile", newReq)
	if newRep.IsSaved {
		log.Printf("The file has been stored at the successor Node: N%s\n", IdentifierGen(SucIP))
	} else {
		log.Println("The file has been not stored on the Ring.")
	}

	return nil
}

//func (cr *ChordRing) AddIdentifier(node ) {
//	cr.Nodes = append(cr.Nodes, node)
//}

func (cr *ChordRing) Notify(NewIPAddress string) bool {
	if cr.Predecessor != "" {
		requestInfo := FindSucRequest{}
		requestInfo.InfoType = GetID
		getPredRep := MakeCall[FindSucRequest, FindSucReplay](cr.Predecessor, "ChordRing.GetNodeInfo", requestInfo) //Call the predecessor node to get its ID(Identifier)
		predID := getPredRep.Identifier                                                                             //Predecessor ID

		getNodeID := MakeCall[FindSucRequest, FindSucReplay](NewIPAddress, "ChordRing.GetNodeInfo", requestInfo)
		newNodeID := getNodeID.Identifier

		isBetween := between(predID, newNodeID, cr.Identifier, false)
		if isBetween {
			cr.Predecessor = NewIPAddress
			//fmt.Printf("Predecessor node IP:%v\n", cr.Predecessor)
			//fmt.Printf("newNodeIP:%v\n", NewIPAddress)
			//fmt.Printf("Current Node IP:%v\n", cr.FullAddress)
			return true
		} else {
			return false
		}

	} else {
		//fmt.Printf("cr.Predecessor before assigning:%v\n", cr.Predecessor)

		cr.Predecessor = NewIPAddress

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
	getReq := FindSucRequest{
		Identifier: cr.Identifier,
		IPAddress:  cr.FullAddress,
	}
	//fmt.Println("joinNodeAdd", joinNodeAdd)
	//findSucReply := CallFS(joinNodeAdd, "ChordRing.FindSuccessor", &getReq)
	findSucReply := MakeCall[FindSucRequest, FindSucReplay](joinNodeAdd, "ChordRing.FindSuccessor", getReq)

	notifyReq := NotifyArgs{}
	cr.Successors[0] = findSucReply.SuccAddress
	notifyReq.NewIPAddress = cr.Successors[0]

	//CallNotify(cr.Successors[0], "ChordRing.NotifyRPC", &notifyReq)
	MakeCall[NotifyArgs, NotifyReply](cr.Successors[0], "ChordRing.NotifyRPC", notifyReq)

	return nil
}

func (cr *ChordRing) Stabilize() error {

	//1. Check on the Successor's predecessor pointer whether it is point back to the current node or not.
	//By calling the successor's predecessor node
	//*//
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	newReq := FindSucRequest{}
	newReq.InfoType = GetPre // Get the Successor's predecessor
	preReplay, err := CallStabilize(cr.Successors[0], "ChordRing.GetNodeInfo", &newReq)
	//fmt.Printf("The predeceussor node of the successor: %v\n", preReplay.Predecessor)
	if err == nil && preReplay.Predecessor != "" {
		newReq.InfoType = GetID

		sucReplay, _ := CallStabilize(preReplay.Predecessor, "ChordRing.GetNodeInfo", &newReq)

		sucId := IdentifierGen(cr.Successors[0]) //Extract the successor's ID from its ip address
		sucPred := sucReplay.Identifier          //Successor's predecessor ID
		isBetween := between(cr.Identifier, sucPred, sucId, false)
		//fmt.Printf("isBetween the predecessor and the current node's successor: %v\n", isBetween)
		if isBetween {
			cr.Successors[0] = preReplay.Predecessor
			//fmt.Printf("New successor: %s, Current Node: %s\n", cr.Successors[0], cr.FullAddress)
		}
	}
	//Notify the successor about its predecessor
	notifyReq := NotifyArgs{}
	notifyReq.NewIPAddress = cr.FullAddress
	//CallNotify(cr.Successors[0], "ChordRing.NotifyRPC", &notifyReq)
	MakeCall[NotifyArgs, NotifyReply](cr.Successors[0], "ChordRing.NotifyRPC", notifyReq)

	return nil
}

func (cr *ChordRing) Check_predecessor() error {
	if cr.Predecessor != "" {
		_, err := rpc.Dial("tcp", cr.Predecessor)
		if err != nil {
			fmt.Printf("The Predecessor:%s of the node: %s is nolonger avalaible\n", cr.Predecessor, cr.FullAddress)
			cr.Predecessor = ""
		}
	}
	return nil
}
