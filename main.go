package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	udp "github.com/TTK4145/Network-go/udp"
	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
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

	send_ch := make(chan udp.Udp_message)
	rcv_ch := make(chan udp.Udp_message)

	/*   NETWORK TEST */
	udp.Udp_init(13378, 13379, 256, send_ch, rcv_ch)

	log.Text("Sending")
	send_ch <- udp.Udp_message{Raddr: "broadcast", Data: "Abdi", Length: 4}

	log.Text("Receiving")
	rcv_msg := <-rcv_ch
	fmt.Printf("msg:  \n \t raddr = %s \n \t data = %s \n \t length = %v \n", rcv_msg.Raddr, rcv_msg.Data, rcv_msg.Length)

	/* END NETWORK TEST */

	// Main loop, will poll for events and act thereafter
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
