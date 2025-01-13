package main

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/rpc"
	"os"
)
import "strconv"

// example to show how to declare the arguments
// and reply for an RPC.
type taskType int
type currentStatus int

// const (
//
//	WorkerIdle = iota
//	WorkerFinishMap
//	WorkerFinishReduce
//
// )
//
// const (
//
//	MapTask = iota
//	ReduceTask
//	WaitForTask
//	NoTask
//
// )
type InfoType int
type StoringType int

// ReqArgs Request type arguments
type ReqArgs struct {
	//CurrentStatus currentStatus
	Key *big.Int
	//FileNames []string
	//intermediateFiles []string
	//ID int
}

const (
	GetIP = iota
	GetID
	GetSuc
	GetPre
	GetSuccessors
	GetPubKey
	GetPriKey
)

const (
	MigrateNode = iota
	MigrateUpload
)

// Replay or respond type from the coordinator
type Replay struct {
	TaskType   taskType //The type of the assigned task from the coordinator
	ID         int      // Task ID
	InputFiles []string // Input files for the task
	NReduce    int      // Number of reduce tasks (for partitioning)
}

type FindSucReplay struct {
	SuccAddress string
	Identifier  *big.Int
	Successor   string
	Predecessor string
	Successors  []string
	AllInfo     *ChordRing
	PublicKey   *rsa.PublicKey
	PrivateKey  *rsa.PrivateKey
}
type FindSucRequest struct {
	InfoType   InfoType
	Identifier *big.Int
	IPAddress  string
}

type NotifyArgs struct {
	InfoType   InfoType
	NewAddress string
	Identifier *big.Int
}

type NotifyReply struct {
	IPAddress  string
	Identifier *big.Int
	isComplete bool
}

type StabilizeArgs struct {
	InfoType     InfoType
	NewIPAddress string
	Identifier   *big.Int
}

type StabilizeReplay struct {
	IPAddress  string
	Identifier *big.Int
	isComplete bool
}
type StoreFileArgs struct {
	StoreType      StoringType
	Key            *big.Int
	FileName       string
	PrevNodeID     *big.Int //Previous Node ID
	Content        []byte
	MigratedBucket map[*big.Int]string
	isStored       bool
	FileContent    map[*big.Int][]byte
}

type StoreFileReply struct {
	IsSaved bool
}

type BackupArgs struct {
	IPAddress string
	Bucket    map[*big.Int]string
}

type BackupReply struct {
	IsSaved bool
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

// MakeCall Calls the call function to conduct a rpc call.
// This function is generic meaning it can take any struct type possible.
func MakeCall[ArgsType any, ReplayType any](NodeAddress string, callFunc string, args ArgsType) ReplayType {
	var replay ReplayType

	makeCall := call(NodeAddress, callFunc, args, &replay)

	if !makeCall {
		fmt.Printf("Failed to call: %s, at node's IP address of: %s!\n", callFunc, NodeAddress)
	}
	return replay
}

// // CallNotify is dedicated to serve the Notify procedure
//
//	func CallNotify(NodeAddress string, callFunc string, args *NotifyArgs) NotifyReply {
//		replay := NotifyReply{}
//
//		makeCall := call(NodeAddress, callFunc, &args, &replay)
//
//		if !makeCall {
//			fmt.Printf("Failed to call: %d, at node's IP address of: %s!\n", callFunc, NodeAddress)
//		}
//		return replay
//	}

func (cr *ChordRing) CallStore(NodeAddress string, callFunc string, args *StoreFileArgs) StoreFileReply {
	replay := StoreFileReply{}
	ListObjectMethods(cr)
	makeCall := call(NodeAddress, callFunc, &args, &replay)

	if !makeCall {
		fmt.Printf("Failed to call: %d, at node's IP address of: %s!\n", callFunc, NodeAddress)
	}
	return replay
}

func CallStabilize(NodeAddress string, callFunc string, args *FindSucRequest) (FindSucReplay, error) {
	replay := FindSucReplay{}
	makeCall := call(NodeAddress, callFunc, &args, &replay)

	if !makeCall {
		return replay, errors.New("Failed to call the node (At the Stabilize stage)")
	}
	return replay, nil
}
func call(address string, rpcname string, args interface{}, reply interface{}) bool {
	//fmt.Println("call", address)
	c, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Printf("dialing:", err)
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)

	if err != nil {
		log.Println("Call service error: ", err)
		return false
	}

	return true
}
