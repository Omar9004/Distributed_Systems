package main

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Key string

type NodeAddress string

// const m = 6
type finger struct {
	Identifier *big.Int
	IPAddress  string
}
type Node struct {
	Identifier  *big.Int
	FingerTable []finger
	Predecessor string
	Successors  []string
	IPAddress   string
	FullAddress string //IP address and Channel Port
	Bucket      map[Key]string
}

//// CreateRing initializes the Chord Ring
//func (n *Node) CreateRing() {
//	n.Predecessor = nil
//	n.Successors = n
//	n.FingerTable = make([]*Node, m)
//}

//func (n *Node) Join(current *Node) {
//	n.Successors = current
//}

func (n *Node) ParseIP(args *InputArgs) {
	if args.IpAddr == "localhost" || args.IpAddr == "127.0.0.1" {
		args.IpAddr = "127.0.0.1"
		n.IPAddress = args.IpAddr

	} else if args.IpAddr == "0.0.0.0" {
		getLocalIp := getLocalAddress()
		n.IPAddress = getLocalIp
	} else {
		getLocalIp := GetPublicIP()
		n.IPAddress = getLocalIp
	}
}

// IdentifierGen Generate an identifier for a given IP address
func IdentifierGen(IPAdd string) *big.Int {
	var identifier *big.Int
	identifier = hashString(IPAdd)
	identifier.Mod(identifier, big.NewInt(int64(math.Pow(2, m))))
	return identifier
}
func (n *Node) NewNode(args *InputArgs) {

	n.ParseIP(args)
	//Merge node's ip address and port into one variable
	IpPort := fmt.Sprintf("%s:%d", n.IPAddress, args.Port)
	n.FullAddress = IpPort

	//Initializing node's Identifier
	n.Identifier = IdentifierGen(n.IPAddress)

	//Initializing node's FingerTable
	n.FingerTable = make([]finger, m+1)
	n.FingerTableInit()

	//Initializing node Predecessor and a list of Successors
	n.Predecessor = ""
	n.Successors = make([]string, args.SuccessorNum)
	n.ScuccessorInit()
}

func (n *Node) ScuccessorInit() {
	for i := 0; i < len(n.Successors); i++ {
		n.Successors[i] = ""
	}
}

// FingerTableInit initializes node n's FingerTable based on the formula of successor = (n.Identifier+ 2^(i-1)) mod 2^m
// Where i'th represents a finger in the table. In page 4 section IV.D 1 =< i =< m (m=6)
func (n *Node) FingerTableInit() {
	n.FingerTable[0].Identifier = n.Identifier
	n.FingerTable[0].IPAddress = n.IPAddress
	for i := 1; i < len(n.FingerTable); i++ {
		addPart := new(big.Int).Add(n.Identifier, big.NewInt(int64(math.Pow(2, float64(i-1))))) // Addition part
		addPart.Mod(addPart, big.NewInt(int64(math.Pow(2, m))))
		n.FingerTable[i].Identifier = addPart
		n.FingerTable[i].IPAddress = n.IPAddress
	}
}

func (n *Node) NodeServer() {
	fmt.Printf("Ip address: %s\n", n.FullAddress)
	rpc.Register(n)
	listener, err := net.Listen("tcp", n.FullAddress)
	if err != nil {
		//fmt.Printf("Error listening on %s: %s\n", IpAddress, err)
		log.Fatalf("Error listening on %s: %s\n", n.FullAddress, err)
	}
	//defer listener.Close()

	go func(listener net.Listener) {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %s\n", err)
				continue
			}

			go func(conn net.Conn) {
				defer conn.Close() // Close the connection after serving
				jsonrpc.ServeConn(conn)
			}(conn)

		}
	}(listener)
	//select {}
}
