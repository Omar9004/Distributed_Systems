package mr

import (
	"fmt"
	"log"
	"sync"
	"time"
	"net"
	"os"
	"net/rpc"
	"net/http"
)

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

type MapReduceTask interface {
	AssignTask(workerId string)
}

func (t *Task) AssignTask(workerId string) {
	t.Status = IN_PROGRESS
	t.WorkerId = workerId
	t.StartedAt = time.Now()
}
   
// Finds the first idle map task then first idle reduce task.
// If task found, then update the task with the worker assignment
// If no tasks, then reply is set to nil
func (c *Coordinator) AssignTask(workerId string, task *MapReduceTask) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
   
	if c.mapTasksRemaining != 0 {
	 for _, mt := range c.mapTasks {
	  if mt.Status == IDLE {
	   fmt.Printf("[Coordinator] Assigning Map Task: %v\n", mt.FileName)
	   mt.AssignTask(workerId)
	   *task = mt
	   break
	  }
	 }
	 return nil
	}
   
	if c.reduceTasksRemaining != 0 {
	 for _, rt := range c.reduceTasks {
	  if rt.Status == IDLE {
	   fmt.Printf("[Coordinator] Assigning Reduce Task: %v\n", rt.Region)
	   rt.AssignTask(workerId)
	   *task = rt
	   break
	  }
	 }
	 return nil
	}
	fmt.Println("[Coordinator] No idle tasks found")
	task = nil
	return nil
   }

// Your code here -- RPC handlers for the worker to call. 
func (c *coordinator) OurRPChdr(args *reqArgs, reply *Replay) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	switch args.taskType{

	case NoTask:
	
	case mapTask:

	case reduceTask:

	case waitForTask:
	}
} 


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
	c.mutex = sync.Mutex{}//for avoiding two workers see a task as idle and begin work on it.

	//Initialize map tasks
	for i, file := range files {
		c.mapTasks[i] = &mapT{
			filename: file,
			nR:       nReduce,
			task:     task{status: free},
		}

	}

	 //Initialize reduce tasks
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
