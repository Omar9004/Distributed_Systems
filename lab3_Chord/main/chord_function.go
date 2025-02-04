package main

import (
	"crypto/rsa"
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
	KeyBackup   map[*big.Int]string
	//certificate x509.CertPool
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
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

func (cr *ChordRing) StoreRPC(key *big.Int, SucIP string, fileName string) error {
	newReq := StoreFileArgs{FileName: fileName, Key: key, StoreType: MigrateUpload}
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

// Notify updates the predecessor status of the ChordRing after the joining of the new node.
// If there is already a predecessor for the called node,
// then the function checks on whether the newly joined lies between the called node's predecessor and the called node(local node).
// Otherwise, if there is no predecessor, then newly joined will be assigned directly as the called node's new predecessor.
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
	if cr.Notify(args.NewAddress) {
		//storeReq := BackupArgs{Bucket: cr.Bucket}
		//MakeCall[BackupArgs, BackupReply](cr.FullAddress, "ChordRing.MigrateBucket", storeReq)
		cr.MigrateBucket(args.NewAddress)
	} else {
		replay.isComplete = false

	}
	return nil
}

func (cr *ChordRing) JoinChord(joinNodeAdd string, args *FindSucRequest, replay *FindSucReplay) error {
	getReq := FindSucRequest{
		Identifier: cr.Identifier,
		IPAddress:  cr.FullAddress,
	}
	findSucReply := MakeCall[FindSucRequest, FindSucReplay](joinNodeAdd, "ChordRing.FindSuccessor", getReq)

	notifyReq := NotifyArgs{}
	cr.Successors[0] = findSucReply.SuccAddress
	notifyReq.NewAddress = cr.FullAddress

	//CallNotify(cr.Successors[0], "ChordRing.NotifyRPC", &notifyReq)
	MakeCall[NotifyArgs, NotifyReply](cr.Successors[0], "ChordRing.NotifyRPC", notifyReq)

	return nil
}
