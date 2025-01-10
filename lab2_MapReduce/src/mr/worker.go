package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type TimerSetup struct {
	sleepTime time.Duration
	ticker    time.Ticker
	quit      chan bool
}

func (t *TimerSetup) Run(task func()) {
	t.ticker = *time.NewTicker(t.sleepTime)
	go func() {
		for {
			select {
			case <-t.ticker.C:
				go task()
			case <-t.quit:
				t.ticker.Stop()
				return
			}
		}
	}()
}

//type ConnectionType struct {
//	CoordinatorIP string
//	WorkerPort    string
//}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string, coordinatorAddress string) {
	//c := ConnectionType{CoordinatorIP: coordinatorAddress,
	//	WorkerPort: workerPort}
	replay := OurCall("Coordinator.RPCHandler", &ReqArgs{CurrentStatus: WorkerIdle}, coordinatorAddress)
	coordinatorCheck := TimerSetup{sleepTime: time.Duration(1000) * time.Millisecond, quit: make(chan bool)}
	coordinatorCheck.WorkerServer()
	coordinatorCheck.Run(func() { coordinatorCheck.Coordinator_check(coordinatorAddress) })

	for {

		replay = OurCall("Coordinator.RPCHandler", &ReqArgs{CurrentStatus: WorkerIdle}, coordinatorAddress)

		switch replay.TaskType {
		case MapTask:
			MapFunction(replay, ReqArgs{}, mapf, coordinatorAddress)
		case ReduceTask:
			ReduceFunction(replay, ReqArgs{}, reducef, coordinatorAddress)
		case WaitForTask:
			// Wait briefly before requesting again
			time.Sleep(500 * time.Millisecond)
			continue
		case NoTask:
			coordinatorCheck.quit <- true
			os.Exit(0)
		default:
			log.Fatalf("Unknown task type: %d\n", replay.TaskType)
		}
	}
}

// Partition key-value pairs into intermediate buckets using ihash to compute the reduceID.
// The reduceID determines which bucket (corresponding to a Reduce task) the key-value pair belongs to.
// This organization ensures that all key-value pairs with the same key are assigned to the same Reduce task,
// making it easier for the Reduce function to aggregate data efficiently.
func partition(kv []KeyValue, nReduce int) [][]KeyValue {
	intermediates := make([][]KeyValue, nReduce)

	for _, k := range kv {
		reduceId := ihash(k.Key) % nReduce
		intermediates[reduceId] = append(intermediates[reduceId], k)
	}
	return intermediates
}
func MapFunction(replay Replay, request ReqArgs, mapf func(string, string) []KeyValue, address string) {
	//Open the file input assigned from the coordinator
	file, err := os.Open(replay.InputFiles[0])

	if err != nil {
		log.Fatalf("Can't open the file", file)
	}
	//Store the content of the file into the memory
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Can't read the content", file)
	}
	file.Close()
	//Use Map Function exits in mrapps to map the given file, returns a keyValue pair list i.e. ("word","1")
	kv := mapf(replay.InputFiles[0], string(content))
	// Declare intermediates is two-dim. list/array, structured as [][]KeyValue.
	// the outer slice:represents nReduce buckets, the inner slice ([]keyValue) stores key-value pairs
	intermediates := partition(kv, replay.NReduce)

	// Write the content of the intermediates to files, where each file will be encoded as "mr-out-reduceId"

	for reduceId, kvList := range intermediates {
		//Name and create the output file according the reduceId
		outFile := fmt.Sprintf("mr-%d-%d", replay.ID, reduceId)
		//file, err := os.Create(outFile)
		file, err := ioutil.TempFile("", outFile)
		request.FileNames = append(request.FileNames, outFile)
		if err != nil {
			log.Fatalf("Failed to create file %s: %v:", outFile, err)
		}

		//Use JSON encoder to write keyValue pair content to a file
		encoder := json.NewEncoder(file)
		for _, kv := range kvList {
			if err := encoder.Encode(kv); err != nil {
				log.Fatalf("Failed to write keyValue pair to file %v: %v", outFile, err)
			}
		}
		os.Rename(file.Name(), outFile)
		file.Close()

	}
	request.ID = replay.ID
	request.CurrentStatus = WorkerFinishMap
	//fmt.Printf("Worker request ar	go coordinatorCheck.WorkerServer()

	// Notify the coordinator that the Map task is completed
	OurCall("Coordinator.TaskComplete", &request, address)
	//newReplay := OurCall("Coordinator.RPCHandler", &request)

}
func ReduceFunction(replay Replay, request ReqArgs, reducef func(string, []string) string, address string) {
	// Create a new list to store all intermediate key-value pairs assigned by the coordinator
	var intermediate []KeyValue
	//fmt.Printf("Intermediates file assigned to Reduce Task function %v", replay.InputFiles)
	//Iterate through the set of reduced files assigned by Reduce task
	for i := 0; i < len(replay.InputFiles); i++ {
		file, err := os.Open(replay.InputFiles[i])
		if err != nil {
			log.Fatalf("Can't open the file %s: %v:", file, err)
		}
		// Decode the content of the file (JSON-encoded key-value pairs)
		decoder := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := decoder.Decode(&kv); err != nil {
				break // Stop decoding when EOF or error occurs
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}
	sort.Slice(intermediate, func(i, j int) bool {
		return intermediate[i].Key < intermediate[j].Key
	})
	oName := fmt.Sprintf("mr-out-%d", replay.ID)
	request.FileNames = append(request.FileNames, oName)
	ofile, _ := ioutil.TempFile("", oName)
	//ofile, _ := os.Create(oName)

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	ofile.Close()

	os.Rename(ofile.Name(), oName)
	request.ID = replay.ID
	request.CurrentStatus = WorkerFinishReduce
	// Notify the coordinator that the Reduce task is completed
	OurCall("Coordinator.TaskComplete", &request, address)
	//newReplay := OurCall("Coordinator.RPCHandler", &request)

}

func OurCall(callFunc string, args *ReqArgs, coordinatorAddress string) Replay {
	replay := Replay{}

	makeCall := call(callFunc, &args, &replay, coordinatorAddress)

	if !makeCall {
		fmt.Printf("call failed!\n")
	}
	return replay
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}, address string) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	//sockname := coordinatorSock()
	c, err := rpc.DialHTTP("tcp", address)

	if err != nil {
		log.Printf("dialing:", err)
		return false
	}
	defer c.Close()
	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	return false
}

func (t *TimerSetup) WorkerServer() {
	//rpc.Register(t)
	//rpc.HandleHTTP()
	l, err := net.Listen("tcp", "192.168.0.106:"+"8080")
	fmt.Printf("The worker is listning %s\n", l.Addr().String())
	if err != nil {
		log.Fatalf("Server is not running", err)
	}
	go http.Serve(l, nil)
}

func (t *TimerSetup) Coordinator_check(CoorAddress string) error {
	_, err := rpc.DialHTTP("tcp", CoorAddress)

	if err != nil {
		log.Println("Coordinator is nolonger avaliable !!")

	}
	return nil
}

//func isEmpty(r Replay) bool {
//	return r.ID == 0 && len(r.InputFiles) == 0 && r.NReduce == 0
//}
