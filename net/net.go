package net

import (
	"strconv"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net/udp"
	crc8 "github.com/mewpkg/hashutil/crc8"
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
 * 1 char: hex-encoded crc8
 */

type OrderType string

const (
	InvalidOrder   OrderType = "INVL"
	NewOrder                 = "NWOD"
	AcceptedOrder            = "ACOD"
	CompletedOrder           = "COOD"
)

type OrderMessage struct {
	Type      OrderType
	Floor     driver.Floor
	Direction driver.ButtonType
}

func orderToStr(order OrderMessage) string {
	//
	str := string(order.Type) + strconv.Itoa(int(order.Floor)) + strconv.Itoa(int(order.Direction))
	crc := crc8.ChecksumATM([]byte(str))
	// This is derp, due to FormatUint not padding 0
	str += strconv.FormatUint(uint64((crc>>4)&0xf), 16) + strconv.FormatUint(uint64(crc&0xf), 16)
	return str
}

func strToOrder(str string) OrderMessage {
	crc := crc8.ChecksumATM([]byte(str[:6]))
	senderCrc, err := strconv.ParseUint(string(str[6:]), 16, 8)

	if err != nil {
		log.Error(err.Error())
	}

	// This is ugly
	if crc != uint8(senderCrc) {
		log.Error("CRC mismatch: ", str[6:], " vs ", strconv.FormatUint(uint64((crc>>4)&0xf), 16)+strconv.FormatUint(uint64(crc&0xf), 16))
		return OrderMessage{Type: InvalidOrder}
	}

	floorNum, _ := strconv.Atoi(string(str[4]))
	dirNum, _ := strconv.Atoi(string(str[4]))

	return OrderMessage{Type: OrderType(str[:4]), Floor: driver.Floor(floorNum), Direction: driver.ButtonType(dirNum)}

}

const LPORT = 13376
const BPORT = 13377
const MSGLEN = 8

// Handles communication with other elevators
func Handler(send <-chan OrderMessage, receive chan<- OrderMessage) {
	udpSendCh := make(chan udp.Udp_message)
	udpRecvCh := make(chan udp.Udp_message)

	udp.Udp_init(LPORT, BPORT, MSGLEN, udpSendCh, udpRecvCh)

	for {
		select {
		case msg := <-udpRecvCh:
			if msg.Length != 8 { // Disregard messages not 8 in length
				log.Warning("Non-8-byte message received")
				continue
			}

			order := strToOrder(msg.Data)

			// TODO: Disregard messages coming from here

			//switch msg.Data[:4] {
			switch order.Type {
			//case string(NewOrder): // New order
			case NewOrder: // New order
				// TODO: Do something sensible
				dir := "up"
				if string(msg.Data[5]) == "1" {
					dir = "down"
				}
				log.Info("New order: floor ", string(msg.Data[4]), ", ", dir)
				break
				//case string(AcceptedOrder): // Accepted order
			case AcceptedOrder: // Accepted order
				dir := "up"
				if string(msg.Data[5]) == "1" {
					dir = "down"
				}
				log.Info("Accepted order: floor ", string(msg.Data[4]), ", ", dir)
				break
				//case string(CompletedOrder): // Completed order
			case CompletedOrder: // Completed order
				dir := "up"
				if string(msg.Data[5]) == "1" {
					dir = "down"
				}
				log.Info("Completed order: floor ", string(msg.Data[4]), ", ", dir)
			}

		case order := <-send:

			//str := string(order.Type) + strconv.Itoa(int(order.Floor)) + strconv.Itoa(int(order.Direction))
			str := orderToStr(order)
			log.Debug("Sending message: ", str)
			udpSendCh <- udp.Udp_message{Raddr: "broadcast", Data: str}
		}
	}

}

func NewExternalOrder() {

}
