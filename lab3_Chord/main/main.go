package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

func main() {

	arguments := ArgsInit()

	if arguments.InputArgsState == InvalidArgs {
		log.Fatal("Invalid input arguments...!!")
		os.Exit(1)
	} else {

		chord := NewNode(arguments)
		//NewChordRing(&node)
		fmt.Printf("Argumnets %v\n", arguments)
		fmt.Printf("chord %v\n", chord)
		chord.NodeServer()
		//err := rpc.Register(chord)
		//if err != nil {
		//	fmt.Printf("Error with rpc Register %v:\n", err.Error())
		//}
		////rpc.HandleHTTP()
		//
		//// sockname := coordinatorSock()
		//// os.Remove(sockname)
		//addr, err := net.ResolveTCPAddr("tcp", chord.FullAddress)
		//
		//if err != nil {
		//	log.Fatal("Inaccessible IP", err.Error())
		//}
		//listener, e := net.Listen("tcp", addr.String())
		//if e != nil {
		//	log.Fatal("listen error:", e)
		//}
		//fmt.Printf("NodeServer is running at %s\n", chord.FullAddress)
		//defer listener.Close()
		//go func(listener net.Listener) {
		//	for {
		//		conn, err := listener.Accept()
		//		if err != nil {
		//			log.Printf("Error accepting connection: %s\n", err)
		//			continue
		//		}
		//
		//		go jsonrpc.ServeConn(conn)
		//
		//	}
		//}(listener)
		switch arguments.InputArgsState {
		case InvalidArgs:

		case NewChord:
			chord.NewChordRing()
		case JoinChord:
			joinNodeAdd := fmt.Sprintf("%s:%d", arguments.NewIp, arguments.NewPort)
			chord.JoinChord(joinNodeAdd, &FindSucRequest{}, &FindSucReplay{})
		}

		input := bufio.NewReader(os.Stdin)

		for {
			log.Println("Press 'q' to exit...")
			in, _ := input.ReadString('\n')
			in = strings.TrimSpace(in)
			if in == "q" {
				os.Exit(0)
			}
		}

	}

}

func ListObjectMethods(obj interface{}) {
	t := reflect.TypeOf(obj)
	fmt.Printf("Methods of %s:\n", t)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		fmt.Printf("  Method: %s\n", method.Name)
	}
}
