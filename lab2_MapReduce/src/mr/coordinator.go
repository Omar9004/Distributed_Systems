package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	//"os"
	"sync"
	"time"
)

// Task statuses
const (
	TaskAvailable = iota
	TaskInProgress
	TaskCompleted
)

// Task type constants

// Task Common task structure
type Task struct {
	Status    int
	WorkerId  int
	StartTime time.Time
}

// MapT Map task structure
type MapT struct {
	Filename string
	NReduce  int
	Task
}

// ReduceT Reduce task structure
type ReduceT struct {
	Place    int
	Location []string
	Task
}

// Coordinator structure
type Coordinator struct {
	//Map task instances
	MapTasks          []*MapT
	MapTasksRemaining int
	MapTaskDone       bool
	//Map task instances
	ReduceTasks          []*ReduceT
	ReduceTasksRemaining int
	ReduceTaskDone       bool
	//Hold the max time duration for each task to do the assigned job
	MaxTaskDuration time.Duration
	// To manage the threads allocation of resources
	Mutex sync.Mutex
}

// TaskComplete A Worker will use this function to notify the Coordinator that the assigned task is completed
func (c *Coordinator) TaskComplete(args *ReqArgs, replay *Replay) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	switch args.CurrentStatus {
	case WorkerFinishMap:
		task := c.MapTasks[args.ID]
		if task.Status == TaskInProgress {
			task.Status = TaskCompleted
			c.MapTasksRemaining--
			if c.MapTasksRemaining == 0 {
				c.MapTaskDone = true
			}

		}
		for reduceId, file := range args.FileNames {
			task := c.ReduceTasks[reduceId]
			task.Location = append(task.Location, file)
		}

	case WorkerFinishReduce:
		task := c.ReduceTasks[args.ID]
		if task.Status == TaskInProgress {
			task.Status = TaskCompleted
			c.ReduceTasksRemaining--
			if c.ReduceTasksRemaining == 0 {
				c.ReduceTaskDone = true
			}

		}
	default:
		log.Fatalf("[Coordinator.taskComplete] unknown task status: %v", args.CurrentStatus)
	}

	return nil
}

// RPCHandler RPC handler for worker requests
func (c *Coordinator) RPCHandler(args *ReqArgs, reply *Replay) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Handle Map tasks
	if c.MapTasksRemaining > 0 {
		for id, task := range c.MapTasks {
			if task.Status == TaskInProgress && time.Since(task.StartTime) > c.MaxTaskDuration {
				task.Status = TaskAvailable
			}
			if task.Status == TaskAvailable {
				task.Status = TaskInProgress
				task.StartTime = time.Now()
				reply.TaskType = MapTask
				reply.ID = id
				reply.InputFiles = []string{task.Filename}
				reply.NReduce = task.NReduce
				return nil
			}
		}
	}

	// Transition to Reduce tasks once the MapTaskDone = True
	if c.MapTaskDone {
		for _, task := range c.ReduceTasks {
			if task.Status == TaskInProgress && time.Since(task.StartTime) > c.MaxTaskDuration {
				task.Status = TaskAvailable
			}
			if task.Status == TaskAvailable {
				task.Status = TaskInProgress
				task.StartTime = time.Now()
				reply.TaskType = ReduceTask
				reply.ID = task.Place
				reply.InputFiles = task.Location
				return nil
			}
		}
	}

	// Signal workers to wait for tasks if no tasks are currently available
	if c.MapTaskDone && c.ReduceTaskDone {
		reply.TaskType = NoTask
		return nil
	}
	reply.TaskType = WaitForTask
	return nil
}

// Start the RPC server
func (c *Coordinator) server(address string) {
	err := rpc.Register(c)
	if err != nil {
		fmt.Printf("Error with rpc Register %v:\n", err.Error())
	}
	rpc.HandleHTTP()

	// sockname := coordinatorSock()
	// os.Remove(sockname)
	l, e := net.Listen("tcp", address)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	fmt.Printf("Coordinator is running at %s\n", address)
	go http.Serve(l, nil)
}

// Done Check if all tasks are completed
func (c *Coordinator) Done() bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.MapTaskDone && c.ReduceTaskDone
}

// MakeCoordinator Create a Coordinator
func MakeCoordinator(files []string, nReduce int, address string) *Coordinator {
	c := Coordinator{
		MapTasks:             make([]*MapT, len(files)),
		ReduceTasks:          make([]*ReduceT, nReduce),
		MapTasksRemaining:    len(files),
		ReduceTasksRemaining: nReduce,
		MapTaskDone:          false,
		ReduceTaskDone:       false,
		MaxTaskDuration:      time.Second * 10,
	}
	// Initialize Map tasks
	for i, file := range files {
		c.MapTasks[i] = &MapT{
			Filename: file,
			NReduce:  nReduce,
			Task: Task{Status: TaskAvailable,
				StartTime: time.Now()},
		}
	}

	// Initialize Reduce tasks
	for i := 0; i < nReduce; i++ {
		c.ReduceTasks[i] = &ReduceT{
			Place: i,
			Task: Task{Status: TaskAvailable,
				StartTime: time.Now()},
		}
	}
	// Start RPC server
	c.server(address)
	return &c
}
