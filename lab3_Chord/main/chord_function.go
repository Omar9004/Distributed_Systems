package main

import (
	"log"
	"math/big"
	//"net/rpc"
	"sync"
	"time"
)

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

	//fmt.Printf("lookupReq.Identifier: %s\n", lookupReq.Identifier)
	//newReplay := CallFS(NodeAddress, "ChordRing.FindSuccessor", &lookupReq)
	newReplay := MakeCall[FindSucRequest, FindSucReplay](NodeAddress, "ChordRing.FindSuccessor", lookupReq)
	return key, newReplay.SuccAddress
}

func (cr *ChordRing) lookupFingers(key *big.Int, NodeAddress string) (*big.Int, string) {

	lookupReq := FindSucRequest{Identifier: key}

	//fmt.Printf("lookupReq.Identifier: %s\n", lookupReq.Identifier)
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
