package mr

import (
	"fmt"
	"log"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

const (
	free = iota
	inProgress
	completed
)

type task struct {
	status    int
	workerId  string
	startTime time.Time
}

type mapT struct {
	filename string
	nR       int
	task
}
type reduceT struct {
	place    int
	location []string
	task
}

type Coordinator struct {
	// Your definitions here.
	mapTasks          []*mapT
	mapTasksRemaining int

	reduceTasks          []*reduceT
	reduceTasksRemaining int

	deadline time.Duration
	mutex    sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("coordinator server listening on", sockname)
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	//ret := false
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// Your code here.
	if c.mapTasksRemaining == 0 && c.reduceTasksRemaining == 0 {
		return true
	}
	return false
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.mapTasks = make([]*mapT, nReduce)
	c.mapTasksRemaining = nReduce
	c.reduceTasks = make([]*reduceT, nReduce)
	c.reduceTasksRemaining = nReduce
	c.deadline = time.Second * 10

	for i, file := range files {
		c.mapTasks[i] = &mapT{
			filename: file,
			nR:       nReduce,
			task:     task{status: free},
		}

	}

	for i := 0; i < nReduce; i++ {
		c.reduceTasks[i] = &reduceT{
			place: i + 1,
			task:  task{status: free},
		}
	}

	fmt.Println("map tasks remaining:", len(c.mapTasks))
	fmt.Println("reduce tasks remaining:", len(c.reduceTasks))

	c.server()
	return &c
}
