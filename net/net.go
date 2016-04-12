package net

import (
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net/udp"
)

/** MESSAGE FORMAT
 * Always 6(TODO: 8) chars
 *
 * 4 chars: type
 * * NWOD = New order
 * * ACOD = Accepted order
 * * COOD = Completed order
 * 1 char: floor (0-indexed)
 * 1 char: direction (0: up, 1: down)
 * TODO: 2 byte crc or whatever
 */

const LPORT = 13376
const BPORT = 13377

// Handles communication with other elevators
func Handler(send <-chan string) {
	sendCh := make(chan udp.Udp_message)
	recvCh := make(chan udp.Udp_message)

	udp.Udp_init(LPORT, BPORT, 256, sendCh, recvCh)

	for {
		select {
		case msg := <-recvCh:
			if msg.Length != 6 { // Disregard messages not 6 in length
				continue
			}

			//log.Debug("Received " + msg.Data[:3] + ", len " + strconv.Itoa(msg.Length) + "!")

			// TODO: Disregard messages coming from here

			switch msg.Data[:4] {
			case "NWOD": // New order
				// TODO: Do something sensible
				dir := "up"
				if string(msg.Data[5]) == "1" {
					dir = "down"
				}
				log.Info("New order: floor ", string(msg.Data[4]), ", ", dir)
				break
			case "ACOD": // Accepted order
				dir := "up"
				if string(msg.Data[5]) == "1" {
					dir = "down"
				}
				log.Info("Accepted order: floor ", string(msg.Data[4]), ", ", dir)
				break
			case "COOD": // Completed order
				dir := "up"
				if string(msg.Data[5]) == "1" {
					dir = "down"
				}
				log.Info("Completed order: floor ", string(msg.Data[4]), ", ", dir)
			}

		case str := <-send:
			log.Debug("Sending message: ", str)
			sendCh <- udp.Udp_message{Raddr: "broadcast", Data: str}
		}
	}

}

func NewExternalOrder() {

}
