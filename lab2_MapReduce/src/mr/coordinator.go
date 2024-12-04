package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
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
	//Map task instances
	ReduceTasks          []*ReduceT
	ReduceTasksRemaining int
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
		}
		//fmt.Printf(" intermediateFiles in the TaskComplete: %v\n", args.intermediateFiles)
		// To save the intermediate files to the location list on Reduce tasks
		for _, file := range args.intermediateFiles {
			var reduceID int

			// Parse the intermediate file name
			sscanf, err := fmt.Sscanf(file, "mr-out-%d-%d", &args.ID, &reduceID)
			if err != nil || sscanf != 2 {
				return fmt.Errorf("failed to parse intermediate file %s: %v", file, err)
			}

			// Lock the Coordinator to safely update ReduceTasks
			//c.Mutex.Lock()
			c.ReduceTasks[reduceID].Location = append(c.ReduceTasks[reduceID].Location, file)
			//c.Mutex.Unlock()

			fmt.Printf("Reduce Task %d updated with intermediate file: %s\n", reduceID, file)
		}
	case WorkerFinishReduce:
		task := c.ReduceTasks[args.ID]
		if task.Status == TaskInProgress {
			task.Status = TaskCompleted
			c.ReduceTasksRemaining--
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
	// Assign a Map task if available
	for _, task := range c.MapTasks {
		if task.Status == TaskInProgress && time.Since(task.StartTime) > c.MaxTaskDuration {
			task.Status = TaskAvailable
			c.MapTasksRemaining++
			break
		}
	}

	// Assign a Map task if available
	for idIndex, task := range c.MapTasks {
		if task.Status == TaskAvailable {
			task.Status = TaskInProgress
			task.StartTime = time.Now()
			reply.TaskType = MapTask
			reply.ID = idIndex
			task.WorkerId = idIndex
			reply.InputFiles = []string{task.Filename}
			reply.NReduce = task.NReduce
			return nil
		}
	}

	for _, task := range c.ReduceTasks {
		if task.Status == TaskInProgress && time.Since(task.StartTime) > c.MaxTaskDuration {
			task.Status = TaskAvailable
			c.ReduceTasksRemaining++
			break
		}
	}
	// If no Map tasks are available, check Reduce tasks
	if c.MapTasksRemaining == 0 {
		for _, task := range c.ReduceTasks {
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

	// No tasks are available
	reply.TaskType = NoTask
	return nil
}

// Start the RPC server
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
	fmt.Println("coordinator server listening on", sockname)
	go http.Serve(l, nil)
}

// Done Check if all tasks are completed
func (c *Coordinator) Done() bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.MapTasksRemaining == 0 && c.ReduceTasksRemaining == 0
}

// MonitorTasks checks if the worker exceeds the time limit
func (c *Coordinator) MonitorTasks() bool {
	for {
		time.Sleep(time.Second)
		c.Mutex.Lock()
		for _, task := range c.MapTasks {
			if task.Status == TaskInProgress && time.Since(task.StartTime) > c.MaxTaskDuration {
				//fmt.Printf("Worker_id: %d is terminated", task.WorkerId)
				task.Status = TaskAvailable
				c.MapTasksRemaining++
			}
		}

		for _, task := range c.ReduceTasks {
			if task.Status == TaskInProgress && time.Since(task.StartTime) > c.MaxTaskDuration {
				task.Status = TaskAvailable
				c.ReduceTasksRemaining++
			}
		}
		c.Mutex.Unlock()
	}
}

// MakeCoordinator Create a Coordinator
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		MapTasks:             make([]*MapT, len(files)),
		ReduceTasks:          make([]*ReduceT, nReduce),
		MapTasksRemaining:    len(files),
		ReduceTasksRemaining: nReduce,
		MaxTaskDuration:      time.Second * 10,
	}
	// Initialize Map tasks
	for i, file := range files {
		c.MapTasks[i] = &MapT{
			Filename: file,
			NReduce:  nReduce,
			Task:     Task{Status: TaskAvailable},
		}
	}

	// Initialize Reduce tasks
	for i := 0; i < nReduce; i++ {
		c.ReduceTasks[i] = &ReduceT{
			Place: i,
			Task:  Task{Status: TaskAvailable},
		}
	}

	fmt.Println("map tasks remaining:", len(c.MapTasks))
	fmt.Println("reduce tasks remaining:", len(c.ReduceTasks))

	// Start RPC server
	c.server()

	// Start monitoring tasks
	go c.MonitorTasks()

	return &c
}
