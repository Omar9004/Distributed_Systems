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
