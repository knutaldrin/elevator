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
	"github.com/knutaldrin/elevator/queue"
)

func main() {
	id := flag.Uint("id", 1337, "Elevator ID")
	flag.Parse()

	if *id > 9 {
		log.Error("Elevator ID must be between 0 and 9")
		os.Exit(1)
	}

	log.Info("Id: ", *id)
	queue.SetID(*id)

	currentDirection := driver.DirectionDown
	lastFloor := driver.Floor(0)

	doorOpen := false

	// Init driver and make sure elevator is at a floor
	driver.Init()

	queue.ImportInternalLog()

	lastFloor = driver.Reset()
	queue.Update(lastFloor)
	queue.ClearOrderLocal(lastFloor, currentDirection)

	floorCh := make(chan driver.Floor)
	go driver.FloorListener(floorCh)

	floorBtnCh := make(chan driver.ButtonEvent, 8)
	go driver.FloorButtonListener(floorBtnCh)

	orderReceiveCh := make(chan net.OrderMessage, 8)
	go net.InitAndHandle(orderReceiveCh, *id)

	timeoutCh := make(chan bool, 8)
	queue.SetTimeoutCh(timeoutCh)

	// Oh, God almighty, please spare our ears
	sigtermCh := make(chan os.Signal)
	signal.Notify(sigtermCh, os.Interrupt, syscall.SIGTERM)
	go func(ch <-chan os.Signal) {
		<-ch
		driver.Stop()
		os.Exit(0)
	}(sigtermCh)

	// Ping timeout so we start in case we have logged orders from a previous crash
	timeoutCh <- true

	// Main event loop
	for {
		select {
		// Elevator has arrived at a new floor
		case fl := <-floorCh:
			queue.Update(fl)
			if queue.ShouldStop(fl) {
				driver.Stop()
				queue.ClearOrderLocal(fl, currentDirection)
				log.Debug("Stopped at floor ", fl)
				net.SendOrder(net.OrderMessage{Type: net.CompletedOrder, Floor: fl, Direction: currentDirection})

				go func() {
					doorOpen = true
					driver.OpenDoor()
					time.Sleep(1 * time.Second)
					doorOpen = false
					driver.CloseDoor()
					timeoutCh <- true
				}()
			}

		// A floor button was pressed
		case btn := <-floorBtnCh:
			queue.NewOrder(btn.Floor, btn.Dir)
			if btn.Dir != driver.DirectionNone {
				net.SendOrder(net.OrderMessage{Type: net.NewOrder, Floor: btn.Floor, Direction: btn.Dir})
			}
			if !doorOpen {
				currentDirection = queue.NextDirection()
				driver.Run(currentDirection)
			}

		// A message came in from the network
		case o := <-orderReceiveCh:
			switch o.Type {
			case net.NewOrder:
				log.Debug("New order, floor: ", o.Floor, ", dir: ", o.Direction)
				queue.NewOrder(o.Floor, o.Direction)

			case net.AcceptedOrder:
				log.Debug("Remote accepted order, floor: ", o.Floor, ", dir: ", o.Direction)
				queue.OrderAcceptedRemotely(o.Floor, o.Direction)

			case net.CompletedOrder:
				log.Debug("Remote completed order, floor: ", o.Floor, ", dir: ", o.Direction)
				queue.ClearOrder(o.Floor, o.Direction)
			}

		// Something timed out. Wake if idle.
		case <-timeoutCh:
			currentDirection = queue.NextDirection()
			if !doorOpen {
				driver.Run(currentDirection)
			}
		}
	}
}
