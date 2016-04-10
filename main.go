package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/knutaldrin/elevator/driver"
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

	/* Fuck this test code shit
	// Start
	if driver.GetFloor() < 3 { */
	driver.RunUp()
	/*} else {
		driver.RunDown()
	}
	*/

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

		case <-floorBtnCh:
			// TODO: Noop
		}

	}
}
