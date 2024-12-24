package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	arguments := ArgsInit()

	if arguments.InputArgsState == InvalidArgs {
		log.Fatal("Invalid input arguments...!!")
		os.Exit(1)
	} else {
		node := &Node{}
		node.NewNode(arguments)
		//NewChordRing(&node)
		fmt.Printf("Argumnets %v\n", arguments)
		go node.NodeServer()

		switch arguments.InputArgsState {
		case InvalidArgs:

		case NewChord:
			NewChordRing(node)
		case JoinChord:

		}
	}

}
