package mr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type ConnectionType struct {
	CoordinatorIP string
	WorkerPort    string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string, coordinatorAddress string, workerPort string) {
	ConnInfo := &ConnectionType{CoordinatorIP: coordinatorAddress, WorkerPort: workerPort}
	go ConnInfo.WorkerServer()
	for {
		replay := OurCall("Coordinator.RPCHandler", &ReqArgs{CurrentStatus: WorkerIdle}, coordinatorAddress)
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
			return
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
	//fmt.Printf("Worker request args: %+v\n", request)

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
		log.Fatal("dialing:", err)
	}
	defer c.Close()
	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func (ConnInfo *ConnectionType) WorkerServer() {
	err := rpc.Register(ConnInfo)
	if err != nil {
		log.Fatalf("Can't register worker: %v", err)
	}
	//if //err != nil {
	//	return
	//}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+ConnInfo.WorkerPort)

	if err != nil {
		log.Fatalf("Server is not running", err)
	}
	go http.Serve(l, nil)
}
