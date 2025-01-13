package main

import (
	"flag"
	"fmt"
)

type InputArgsState int

const (
	InvalidArgs = iota
	NewChord
	JoinChord
)

type InputArgs struct {
	IpAddr           string
	Port             int
	JoinIP           string
	JoinPort         int
	Stabilize        int
	FixFingers       int
	CheckPredecessor int
	SuccessorNum     int
	Identifier       string
	InputArgsState   InputArgsState
}

func ArgsInit() *InputArgs {
	IpAddr := flag.String("a", "localhost", "The IP address that the Chord client will bind to")
	Port := flag.Int("p", 1234, "The port address that the Chord client will bind to")
	JoinIp := flag.String("ja", "xxx", "The Chord client ip that will join this node’s ring")
	JoinPort := flag.Int("jp", 0000, "The Chord client port that will join this node’s ring")
	Stabilize := flag.Int("ts", 3000, "The time between invocations of stabilize")
	FixFingers := flag.Int("tff", 1000, "The time between invocations of fix fingers")
	CheckPredecessor := flag.Int("tcp", 1000, "The time between invocations of check predecessor")
	SuccessorNum := flag.Int("r", 1, "The number of successors maintained by the Chord client")
	Identifier := flag.String("i", "", " The identifier (ID) assigned to the Chord client(Optional)")
	flag.Parse()

	args := InputArgs{
		IpAddr:           *IpAddr,
		Port:             *Port,
		JoinIP:           *JoinIp,
		JoinPort:         *JoinPort,
		Stabilize:        *Stabilize,
		FixFingers:       *FixFingers,
		CheckPredecessor: *CheckPredecessor,
		SuccessorNum:     *SuccessorNum,
		Identifier:       *Identifier,
	}
	if (args.JoinIP == "xxx") && (args.JoinPort == 0000) {
		args.InputArgsState = NewChord
	} else {
		args.InputArgsState = JoinChord
	}
	if args.ArgsValidation() {
		return &args
	}
	//log.Fatal("Invalid arguments")
	args.InputArgsState = InvalidArgs
	return nil
}

func (args *InputArgs) ArgsValidation() bool {
	if args.InputArgsState == NewChord {
		PortCheck := args.Port >= 0 && args.Port <= 65535
		StabilizeCheck := args.Stabilize >= 1 && args.Stabilize <= 60000
		FixFingersCheck := args.FixFingers >= 1 && args.FixFingers <= 60000
		CheckPredecessorC := args.CheckPredecessor >= 1 && args.CheckPredecessor <= 60000
		SuccessorNumCheck := args.SuccessorNum >= 1 && args.SuccessorNum <= 32

		// Log failing conditions
		if !PortCheck {
			fmt.Println("Validation failed: Port must be between 0 and 65535")
		}
		if !StabilizeCheck {
			fmt.Println("Validation failed: Stabilize must be between 1 and 60000 ms")
		}
		if !FixFingersCheck {
			fmt.Println("Validation failed: FixFingers must be between 1 and 60000 ms")
		}
		if !CheckPredecessorC {
			fmt.Println("Validation failed: CheckPredecessor must be between 1 and 60000 ms")
		}
		if !SuccessorNumCheck {
			fmt.Println("Validation failed: SuccessorNum must be between 1 and 32")
		}

		return PortCheck && StabilizeCheck && FixFingersCheck && CheckPredecessorC && SuccessorNumCheck

	} else if args.InputArgsState == JoinChord {
		PortCheck := args.Port >= 0 && args.Port <= 65535 && args.JoinPort >= 0 && args.JoinPort <= 65535
		StabilizeCheck := args.Stabilize >= 1 && args.Stabilize <= 60000
		FixFingersCheck := args.FixFingers >= 1 && args.FixFingers <= 60000
		CheckPredecessorC := args.CheckPredecessor >= 1 && args.CheckPredecessor <= 60000
		SuccessorNumCheck := args.SuccessorNum >= 1 && args.SuccessorNum <= 32

		// Log failing conditions
		if !PortCheck {
			fmt.Println("Validation failed: Ports must be between 0 and 65535")
		}
		if !StabilizeCheck {
			fmt.Println("Validation failed: Stabilize must be between 1 and 60000 ms")
		}
		if !FixFingersCheck {
			fmt.Println("Validation failed: FixFingers must be between 1 and 60000 ms")
		}
		if !CheckPredecessorC {
			fmt.Println("Validation failed: CheckPredecessor must be between 1 and 60000 ms")
		}
		if !SuccessorNumCheck {
			fmt.Println("Validation failed: SuccessorNum must be between 1 and 32")
		}

		return PortCheck && StabilizeCheck && FixFingersCheck && CheckPredecessorC && SuccessorNumCheck
	}

	fmt.Println("Validation failed: Invalid InputArgsState")
	return false
}

func menuPrinter() {
	fmt.Println("Choose the function name by typing it:")
	fmt.Println("1.Lookup")
	fmt.Println("2.StoreFile")
	fmt.Println("3.PrintState")
	fmt.Println("q to Quit")
}
