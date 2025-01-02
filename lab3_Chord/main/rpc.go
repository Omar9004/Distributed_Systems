package main

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
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
}
type FindSucRequest struct {
	InfoType   InfoType
	Identifier *big.Int
	IPAddress  string
}

type NotifyArgs struct {
	InfoType     InfoType
	NewIPAddress string
	Identifier   *big.Int
}

type NotifyReply struct {
	IPAddress  string
	Identifier *big.Int
	isComplete bool
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
func (cr *ChordRing) CallFS(coordinatorAddress string, callFunc string, args *FindSucRequest) FindSucReplay {
	replay := FindSucReplay{}

	makeCall := cr.call(coordinatorAddress, callFunc, &args, &replay)

	if !makeCall {
		fmt.Printf("Failed to call: %d, at node's IP address of: %s!\n", callFunc, coordinatorAddress)
	}
	return replay
}

func (cr *ChordRing) CallNotify(coordinatorAddress string, callFunc string, args *NotifyArgs) NotifyReply {
	replay := NotifyReply{}

	makeCall := cr.call(coordinatorAddress, callFunc, &args, &replay)

	if !makeCall {
		fmt.Printf("Failed to call: %d, at node's IP address of: %s!\n", callFunc, coordinatorAddress)
	}
	return replay
}
func (cr *ChordRing) call(address string, rpcname string, args interface{}, reply interface{}) bool {
	fmt.Println("call", address)
	c, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)

	if err != nil {
		log.Println("Call service error: ", err)
		return false
	}

	return true
}
