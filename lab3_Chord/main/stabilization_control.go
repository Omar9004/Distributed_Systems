package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"time"
)

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
func (cr *ChordRing) initDurations(args *InputArgs) []*TimerSetup {

	timers := []*TimerSetup{
		{sleepTime: time.Duration(args.Stabilize) * time.Millisecond, quit: make(chan bool)},
		{sleepTime: time.Duration(args.FixFingers) * time.Millisecond, quit: make(chan bool)},
		{sleepTime: time.Duration(args.CheckPredecessor) * time.Millisecond, quit: make(chan bool)},
	}

	timers[0].Run(func() { cr.Stabilize() })
	timers[1].Run(func() { cr.FixFingers() })
	timers[2].Run(func() { cr.Check_predecessor() })

	return timers

}

func (cr *ChordRing) Stabilize() error {

	//1. Check on the Successor's predecessor pointer whether it is point back to the current node or not.
	//By calling the successor's predecessor node
	//*//
	//cr.mutex.Lock()
	//defer cr.mutex.Unlock()

	newReq := FindSucRequest{}
	newReq.InfoType = GetPre // Get the Successor's predecessor
	preReplay, err := CallStabilize(cr.Successors[0], "ChordRing.GetNodeInfo", &newReq)
	//fmt.Printf("The predeceussor node of the successor: %v\n", preReplay.Predecessor)
	if err == nil && preReplay.Predecessor != "" {
		newReq.InfoType = GetID

		sucReplay, _ := CallStabilize(preReplay.Predecessor, "ChordRing.GetNodeInfo", &newReq)

		sucId := IdentifierGen(cr.Successors[0]) //Extract the successor's ID from its ip address
		sucPred := sucReplay.Identifier          //Successor's predecessor ID
		isBetween := between(cr.Identifier, sucPred, sucId, false)
		//fmt.Printf("isBetween the predecessor and the current node's successor: %v\n", isBetween)
		if isBetween {
			cr.Successors[0] = preReplay.Predecessor
			//fmt.Printf("New successor: %s, Current Node: %s\n", cr.Successors[0], cr.FullAddress)
		}
	}
	//Notify the successor about its predecessor
	notifyReq := NotifyArgs{}
	notifyReq.NewAddress = cr.FullAddress
	//CallNotify(cr.Successors[0], "ChordRing.NotifyRPC", &notifyReq)
	MakeCall[NotifyArgs, NotifyReply](cr.Successors[0], "ChordRing.NotifyRPC", notifyReq)

	//Step 2: Update the Successor List, in case a node has left the ChordRing
	successorListReq := FindSucRequest{
		InfoType: GetSuccessors,
	}
	succListReply, err := CallStabilize(cr.Successors[0], "ChordRing.GetNodeInfo", &successorListReq)
	if err == nil {
		// Update the local node's Successor list about the latest successors in the Ring
		cr.Successors = append([]string{cr.Successors[0]}, succListReply.Successors...)
		// Maintain the list size equal to SuccessorNum
		if len(cr.Successors) > cr.SuccessorNum {
			cr.Successors = cr.Successors[:cr.SuccessorNum]
		}
		//fmt.Printf("Update successors: %v\n", cr.Successors)
	} else {
		//In case the first Successor is not alive, then the code address the first successor in the list
		fmt.Printf("Failed to update successors: %v\n", err)
		if len(cr.Successors) > 1 {
			cr.Successors = cr.Successors[1:]
		} else {
			// Addressing the case when there is just one node left in Ring,
			// then we assign the successor lists' items to its own local address
			cr.Successors = []string{cr.FullAddress}
		}
	}
	return nil
}

// FixFingers keeps the node's finger_table up-to-date with the latest status of the ChordRing.
func (cr *ChordRing) FixFingers() error {
	//cr.mutex.Lock()
	//defer cr.mutex.Unlock()
	i := rand.Intn(m-1) + 1
	fingerKey := jump(cr.FullAddress, i)

	_, nextNode := cr.lookupFingers(fingerKey, cr.FullAddress)

	cr.FingerTable[i].IPAddress = nextNode
	cr.FingerTable[i].Identifier = fingerKey

	return nil
}
func (cr *ChordRing) Check_predecessor() error {
	if cr.Predecessor != "" {
		_, err := rpc.Dial("tcp", cr.Predecessor)
		if err != nil {
			fmt.Printf("The Predecessor:%s of the node: %s is nolonger avalaible\n", cr.Predecessor, cr.FullAddress)
			cr.Predecessor = ""
		}
	}
	return nil
}
