package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
	"github.com/knutaldrin/elevator/queue"
)

// fuck main
func main() {

	id := flag.Uint("id", 1337, "Elevator ID")

	flag.Parse()

	if *id == 1337 {
		log.Error("Elevator ID must be set")
		os.Exit(1)
	}

	if *id > 9 {
		log.Error("Id plz less than 10")
		os.Exit(1)
	}

	fmt.Println("Id:", *id)
	queue.SetID(*id)

	currentDirection := driver.DirectionNone
	lastFloor := driver.Floor(0)

	doorOpen := false

	// Init driver and make sure elevator is at a floor
	driver.Init()

	queue.ImportInternalLog()

	lastFloor = driver.Reset()
	queue.Update(lastFloor)

	floorCh := make(chan driver.Floor)
	go driver.FloorListener(floorCh)

	stopCh := make(chan bool)
	go driver.StopButtonListener(stopCh)

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
		//panic("CtrlC panic")
		os.Exit(0)
	}(sigtermCh)

	for {
		select {
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
					currentDirection = queue.NextDirection()
					driver.Run(currentDirection)
				}()
			}

		case btn := <-floorBtnCh:
			queue.NewOrder(btn.Floor, btn.Dir)
			if btn.Dir != driver.DirectionNone {
				net.SendOrder(net.OrderMessage{Type: net.NewOrder, Floor: btn.Floor, Direction: btn.Dir})
			}
			if !doorOpen {
				currentDirection = queue.NextDirection()
				driver.Run(currentDirection)
			}

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

		case <-timeoutCh:
			currentDirection = queue.NextDirection()
			if !doorOpen {
				driver.Run(currentDirection)
			}
		}
	}
}
