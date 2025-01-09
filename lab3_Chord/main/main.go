package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
)

func main() {

	arguments := ArgsInit()

	if arguments.InputArgsState == InvalidArgs {
		log.Fatal("Invalid input arguments...!!")
		os.Exit(0)
	} else {
		chord := NewNode(arguments)
		//NewChordRing(&node)
		//fmt.Printf("Argumnets %v\n", arguments)
		//fmt.Printf("chord %v\n", chord)
		//fmt.Printf("%s\n", chord.Identifier)
		chord.NodeServer()
		switch arguments.InputArgsState {
		case InvalidArgs:

		case NewChord:
			chord.NewChordRing()
		case JoinChord:
			joinNodeAdd := fmt.Sprintf("%s:%d", arguments.NewIp, arguments.NewPort)
			chord.JoinChord(joinNodeAdd, &FindSucRequest{}, &FindSucReplay{})
		}
		timers := chord.initDurations(arguments)
		input := bufio.NewReader(os.Stdin)

		for {
			menuPrinter()
			in := readStringIn(input)
			switch in {
			case "Lookup":
				//file, _ := input.ReadString('\n')
				//file = strings.TrimSpace(file)
				file := readStringIn(input)
				SucAddress, _ := chord.lookup(file, chord.FullAddress)
				fmt.Printf("The file exists at the node with an IP address: %s\n", SucAddress)
			case "StoreFile":
				fmt.Println("Enter the name of the file that you want to store!")
				file := readStringIn(input)
				key, SucAddress := chord.lookup(file, chord.FullAddress)
				fmt.Printf("The file exists at the node with an IP address: %s\n", SucAddress)
				chord.storeFile(key, SucAddress, file)
				fmt.Printf("The file exists at the node with an IP address: %s\n", SucAddress)
				fmt.Printf("The file has a key of : %s\n", key)
			case "q":
				if timers != nil {
					timers[0].quit <- true //Inform the Stabilizer about the node exiting to terminate the Goroutine
					timers[1].quit <- true //Inform the FixFingers about the node exiting to terminate the Goroutine
					timers[2].quit <- true //Inform the Check_Predecessor about the node exiting to terminate the Goroutine
				}
				os.Exit(0)
			default:
				fmt.Println("Invalid input")
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
