package main

import "time"

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
	if args.InputArgsState != NewChord {
		timers := []*TimerSetup{
			{sleepTime: time.Duration(args.Stabilize) * time.Millisecond, quit: make(chan bool)},
			{},
			{sleepTime: time.Duration(args.CheckPredecessor) * time.Millisecond, quit: make(chan bool)},
		}

		timers[0].Run(func() { cr.Stabilize() })
		timers[2].Run(func() { cr.Check_predecessor() })

		return timers
	}
	return nil
}
