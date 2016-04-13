package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
)

func main() {

	// Init driver and make sure elevator is at a floor
	driver.Init()
	driver.Reset()

	floorCh := make(chan driver.Floor)
	go driver.FloorListener(floorCh)

	stopCh := make(chan bool)
	go driver.StopButtonListener(stopCh)

	floorBtnCh := make(chan driver.ButtonEvent)
	go driver.FloorButtonListener(floorBtnCh)

	orderSendCh := make(chan net.OrderMessage)
	orderReceiveCh := make(chan net.OrderMessage)
	go net.Handler(orderSendCh, orderReceiveCh)

	driver.RunUp()

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
			// TODO: If correct floor, let queue know AND ->
			// Turn off light and send network message
			driver.ButtonLightOff(fl, driver.DirectionNone)
			switch fl {
			case 0:
				driver.RunUp()
			case 3:
				driver.RunDown()
			}

		case stop := <-stopCh:
			if stop {
				driver.Stop()
			}

		case fl := <-floorBtnCh:
			if fl.Dir == driver.DirectionDown || fl.Dir == driver.DirectionUp {
				// TODO: Let queue know
				// Send network message
				orderSendCh <- net.OrderMessage{Type: net.NewOrder, Floor: fl.Floor, Direction: fl.Dir}
			} else {
				// TODO: Let queue know about the order
				log.Info("Internal order for floor " + strconv.Itoa(int(fl.Floor)))
				driver.ButtonLightOn(fl.Floor, fl.Dir)
			}
		}
	}
}
