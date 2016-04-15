package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
	"github.com/knutaldrin/elevator/newq"
)

// fuck main
func main() {

	id := flag.Uint("id", 1337, "Elevator ID")

	flag.Parse()

	if *id == 1337 {
		log.Error("Elevator ID must be set")
		os.Exit(1)
	}

	currentDirection := driver.DirectionNone
	lastFloor := driver.Floor(0)

	// Init driver and make sure elevator is at a floor
	driver.Init()

	lastFloor = driver.Reset()
	newq.Update(lastFloor)

	floorCh := make(chan driver.Floor)
	go driver.FloorListener(floorCh)

	stopCh := make(chan bool)
	go driver.StopButtonListener(stopCh)

	floorBtnCh := make(chan driver.ButtonEvent)
	go driver.FloorButtonListener(floorBtnCh)

	orderReceiveCh := make(chan net.OrderMessage)
	go net.InitAndHandle(orderReceiveCh)

	timeoutCh := make(chan bool)
	newq.SetTimeoutCh(timeoutCh)

	// Oh, God almighty, please spare our ears
	sigtermCh := make(chan os.Signal)
	signal.Notify(sigtermCh, os.Interrupt, syscall.SIGTERM)
	go func(ch <-chan os.Signal) {
		<-ch
		driver.Stop()
		os.Exit(0)
	}(sigtermCh)

	for {
		select {
		case fl := <-floorCh:
			newq.Update(fl)
			if newq.ShouldStop(fl) {
				driver.Stop()
				newq.ClearOrder(fl, currentDirection)
				log.Debug("Stopped at floor ", fl)
				// TODO: Open doors not blockking, so people can still press buttons
				driver.OpenDoor()
				time.Sleep(1 * time.Second)
				driver.CloseDoor()
			}

		case btn := <-floorBtnCh:
			newq.NewOrder(btn.Floor, btn.Dir)
			driver.Run(newq.NextDirection())

		case <-timeoutCh:
			currentDirection = newq.NextDirection()
			driver.Run(currentDirection)
		}
	}
}
