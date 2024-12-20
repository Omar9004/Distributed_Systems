package lab3_Chord

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)
import "strconv"

// example to show how to declare the arguments
// and reply for an RPC.
type taskType int
type currentStatus int

const (
	WorkerIdle = iota
	WorkerFinishMap
	WorkerFinishReduce
)

const (
	MapTask = iota
	ReduceTask
	WaitForTask
	NoTask
)

// ReqArgs Request type arguments
type ReqArgs struct {
	CurrentStatus currentStatus
	FileNames     []string
	//intermediateFiles []string
	ID int
}

// Replay or respond type from the coordinator
type Replay struct {
	TaskType   taskType //The type of the assigned task from the coordinator
	ID         int      // Task ID
	InputFiles []string // Input files for the task
	NReduce    int      // Number of reduce tasks (for partitioning)
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
func (c *Coordinator) server() {
	err := rpc.Register(c)
	if err != nil {
		fmt.Printf("Error with rpc Register %v:\n", err.Error())
	}
	rpc.HandleHTTP()

	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}
