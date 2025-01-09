package main

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"os"
)

type Key string

type NodeAddress string

// const m = 6

//type Node struct {
//	Identifier  *big.Int
//	FingerTable []finger
//	Predecessor string
//	Successors  []string
//	IPAddress   string
//	FullAddress string //IP address and Channel Port
//	Bucket      map[Key]string
//}

//// CreateRing initializes the Chord Ring
//func (n *Node) CreateRing() {
//	n.Predecessor = nil
//	n.Successors = n
//	n.FingerTable = make([]*Node, m)
//}

//func (n *Node) Join(current *Node) {
//	n.Successors = current
//}

func (cr *ChordRing) ParseIP(args *InputArgs) {
	if args.IpAddr == "localhost" || args.IpAddr == "127.0.0.1" {
		args.IpAddr = "127.0.0.1"
		cr.IPAddress = args.IpAddr

	} else if args.IpAddr == "0.0.0.0" {
		getLocalIp := getLocalAddress()
		cr.IPAddress = getLocalIp
	} else if args.IpAddr == "public" {
		getLocalIp := GetPublicIP()
		cr.IPAddress = getLocalIp
	} else {
		cr.IPAddress = args.IpAddr
	}
}

func NewNode(args *InputArgs) *ChordRing {
	cr := &ChordRing{}
	cr.ParseIP(args)
	//Merge node's ip address and port into one variable
	IpPort := fmt.Sprintf("%s:%d", cr.IPAddress, args.Port)
	cr.FullAddress = IpPort
	//Initializing node's Identifier
	cr.Identifier = IdentifierGen(cr.FullAddress)

	//Initializing node's FingerTable
	cr.FingerTable = make([]finger, m+1)
	cr.FingerTableInit()

	//Initializing node Predecessor and a list of Successors
	cr.Predecessor = ""
	cr.Successors = make([]string, args.SuccessorNum)
	cr.SuccessorInit()

	//Initialize the node's Bucket
	cr.Bucket = make(map[*big.Int]string)

	FolderName := FolderPathGen(cr.Identifier)
	// Create a directory for a node for saving the assigned files on it.
	err := os.MkdirAll(FolderName, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory")
	}
	cr.NodeFolder = FolderName
	return cr
}

func (cr *ChordRing) SuccessorInit() {
	for i := 0; i < len(cr.Successors); i++ {
		cr.Successors[i] = ""
	}
}

// FingerTableInit initializes node n's FingerTable based on the formula of successor = (n.Identifier+ 2^(i-1)) mod 2^m
// Where i'th represents a finger in the table. In page 4 section IV.D 1 =< i =< m (m=6)
func (cr *ChordRing) FingerTableInit() {
	cr.FingerTable[0].Identifier = cr.Identifier
	cr.FingerTable[0].IPAddress = cr.IPAddress
	for i := 1; i < len(cr.FingerTable); i++ {
		addPart := jump(cr.IPAddress, i) //Hashing an IP address for given node
		cr.FingerTable[i].Identifier = addPart
		cr.FingerTable[i].IPAddress = cr.FullAddress
	}
}

func (cr *ChordRing) FindClosetFinger(id *big.Int) string {
	requestInfo := FindSucRequest{}
	requestInfo.InfoType = GetIP
	//fmt.Printf("FingerTable Size: %d\n", len(cr.FingerTable))
	for i := len(cr.FingerTable) - 1; i > 0; i-- {
		//sucReplay := CallFS(cr.FingerTable[i].IPAddress, "ChordRing.GetNodeInfo", &requestInfo)
		sucReplay := MakeCall[FindSucRequest, FindSucReplay](cr.FingerTable[i].IPAddress, "ChordRing.GetNodeInfo", requestInfo)
		SucId := IdentifierGen(sucReplay.SuccAddress)
		isBetween := between(cr.Identifier, SucId.Mod(SucId, hashMod), id, false)
		if isBetween {
			return cr.FingerTable[i].IPAddress
		}

	}
	return cr.Successors[0]
}

func (cr *ChordRing) GetNodeInfo(args *FindSucRequest, replay *FindSucReplay) error {
	switch args.InfoType {
	case GetIP:
		replay.SuccAddress = cr.FullAddress
	case GetID:
		replay.Identifier = cr.Identifier
	case GetSuc:
		replay.Successor = cr.Successors[0]
	case GetPre:
		replay.Predecessor = cr.Predecessor
	case GetSuccessors:
		replay.Successors = cr.Successors
	default:
		return errors.New("invalid info type")
	}
	return nil
}

func (cr *ChordRing) NodeServer() {
	err := rpc.Register(cr)
	if err != nil {
		fmt.Printf("Error with rpc Register %v:\n", err.Error())
	}
	//rpc.HandleHTTP()

	// sockname := coordinatorSock()
	// os.Remove(sockname)
	addr, err := net.ResolveTCPAddr("tcp", cr.FullAddress)

	if err != nil {
		log.Fatal("Inaccessible IP", err.Error())
	}
	listener, e := net.Listen("tcp", addr.String())
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Printf("NodeServer is running at %s\n", cr.FullAddress)
	go func(listener net.Listener) {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %s\n", err)
				break
			}
			go rpc.ServeConn(conn)

		}
	}(listener)
}

func (cr *ChordRing) StoreFile(args *StoreFileArgs, replay *StoreFileReply) error {
	key := args.Key

	//if file exists on the current directory /main, then move it to the node's folder
	if _, err := os.Stat(args.FileName); err == nil {
		nodeFolder := cr.NodeFolder
		err := os.Rename(args.FileName, nodeFolder+"/"+args.FileName)
		if err != nil {
			fmt.Printf("Error moving file to %s", nodeFolder+"/"+args.FileName)
			replay.IsSaved = false
		}
		cr.Bucket[key] = args.FileName

	} else {
		cr.Bucket[key] = args.FileName
		replay.IsSaved = true
	}
	return nil
}

func (cr *ChordRing) FindSuccessor(args *FindSucRequest, replay *FindSucReplay) error {
	//fmt.Printf("Joined Node ip: %s\n", args.IPAddress)
	//fmt.Printf("Joined Node id: %s\n", args.Identifier)
	//cr.mutex.Lock()
	//defer cr.mutex.Unlock()
	requestInfo := FindSucRequest{}
	requestInfo.InfoType = GetIP
	joinNodeID := args.Identifier.Mod(args.Identifier, hashMod)
	//newReplay := CallFS(cr.Successors[0], "ChordRing.GetNodeInfo", &requestInfo)
	newReplay := MakeCall[FindSucRequest, FindSucReplay](cr.Successors[0], "ChordRing.GetNodeInfo", requestInfo)

	//if !cr.call(cr.Successors[0], "ChordRing.GetSucId", FindSucRequest{}, &newReplay) {
	//	return fmt.Errorf("Faild to reach GetSucID %v\n", newReplay)
	//}
	idSuc := IdentifierGen(newReplay.SuccAddress)
	fmt.Println(args.Identifier)
	isBetween := between(cr.Identifier, joinNodeID, idSuc, true)
	if isBetween {
		//fmt.Printf("Found successor in between: %s\n", args.IPAddress)

		replay.SuccAddress = newReplay.SuccAddress

	} else { //Otherwise search on the finger table of this node
		sucAddress := cr.FindClosetFinger(idSuc)
		//sucReplay := CallFS(sucAddress, "ChordRing.FindSuccessor", args)
		newArgs := FindSucRequest{Identifier: joinNodeID}
		sucReplay := MakeCall[FindSucRequest, FindSucReplay](sucAddress, "ChordRing.FindSuccessor", newArgs)

		replay.SuccAddress = sucReplay.SuccAddress
	}
	//fmt.Printf("Found the successor: %s\n", replay.SuccAddress)
	//fmt.Printf("Found suc between %s\n", isBetween)
	return nil
}
