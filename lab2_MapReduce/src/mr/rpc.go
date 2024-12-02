package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

// example to show how to declare the arguments
// and reply for an RPC.
type taskType int
type currentStatus int

const (
	workerIdle = iota
	workerBusy
	workerFinishMap
	workerFinishReduce
)

const (
	mapTask = iota
	reduceTask
	waitForTask
	noTask
)

// Request type arguments
type reqArgs struct {
	currentStatus currentStatus
	fileNames     []string
	id            int
}

// Replay or respond type from the coordinator
type Replay struct {
	taskType taskType //The type of the assigned task from the coordinator
	id       int      // Task ID
	files    []string // Input files for the task
	nReduce  int      // Number of reduce tasks (for partitioning)
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
