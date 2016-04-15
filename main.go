package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
	"github.com/knutaldrin/elevator/queue"
)

//Utilities
func goUp() {
	driver.RunUp()
	queue.SetDirStatus(driver.DirectionUp)
}

func goDown() {
	driver.RunDown()
	queue.SetDirStatus(driver.DirectionDown)
}

func stop() {
	driver.Stop()
	queue.SetDirStatus(driver.DirectionNone)
}

// doorPiccolo should be spawned as a goroutine and handles stopping and opening the door
func doorPiccolo(nextDirCh chan<- driver.Direction) {
	//ButtonLightOff(fl, DirectionNone)
	log.Info("Stopped at floor")
	driver.OpenDoor()
	log.Info("Door open")
	time.Sleep(2 * time.Second)
	driver.CloseDoor()
	log.Info("Door closed")
	queue.NextDir(nextDirCh)
}

//MAAAAAAAAAAAAAAAAAAAAAAIN
func main() {

	id := flag.Uint("id", 1337, "Elevator ID")

	flag.Parse()

	if *id == 1337 {
		log.Error("Elevator ID must be set")
		os.Exit(1)
	}

	// Init driver and make sure elevator is at a floor
	driver.Init()
	queue.ShouldStopAtFloor(driver.Reset())

	floorCh := make(chan driver.Floor)
	go driver.FloorListener(floorCh)

	stopCh := make(chan bool)
	go driver.StopButtonListener(stopCh)

	floorBtnCh := make(chan driver.ButtonEvent)
	go driver.FloorButtonListener(floorBtnCh)

	orderSendCh := make(chan net.OrderMessage)
	orderReceiveCh := make(chan net.OrderMessage)
	go net.Handler(orderSendCh, orderReceiveCh)

	orderToQueueCh := make(chan net.OrderMessage)
	go queue.Manager(orderToQueueCh, *id)

	nextDirCh := make(chan driver.Direction)

	// Oh, God almighty, please spare our ears
	sigtermCh := make(chan os.Signal)
	signal.Notify(sigtermCh, os.Interrupt, syscall.SIGTERM)
	go func(ch <-chan os.Signal) {
		<-ch
		driver.Stop()
		os.Exit(0)
	}(sigtermCh)

	go queue.NextDir(nextDirCh)

	for {
		log.Bullshit("Selecting")
		select {
		case fl := <-floorCh:
			if queue.ShouldStopAtFloor(fl) {
				stop()
				go doorPiccolo(nextDirCh)
			}

		case btn := <-floorBtnCh:
			if btn.Dir == driver.DirectionDown || btn.Dir == driver.DirectionUp {
				// TODO: Let queue know
				// Send network message
				orderSendCh <- net.OrderMessage{Type: net.NewOrder, Floor: btn.Floor, Direction: btn.Dir}
			} else {
				orderToQueueCh <- net.OrderMessage{Type: net.InternalOrder, Floor: btn.Floor, Direction: driver.DirectionNone}
				log.Info("Internal order for floor " + strconv.Itoa(int(btn.Floor)))
				driver.ButtonLightOn(btn.Floor, btn.Dir)
			}

		case dir := <-nextDirCh:
			switch dir {
			case driver.DirectionUp:
				goUp()
			case driver.DirectionDown:
				goDown()
			case driver.DirectionNone:
				stop()
				go doorPiccolo(nextDirCh)
			}
		}
	}
}
