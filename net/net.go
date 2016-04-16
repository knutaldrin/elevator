package net

import (
	"strconv"

	crc16 "github.com/joaojeronimo/go-crc16"
	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net/udp"
)

/** MESSAGE FORMAT
 * Always 8 chars
 *
 * 4 chars: type
 * * NWOD = New order
 * * ACOD = Accepted order
 * * COOD = Completed order
 * 1 char: floor (0-indexed)
 * 1 char: direction (0: up, 1: down)
 * 2 chars: CRC-16 of the previous 6 bytes
 */

//OrderType is an enum for communicating information about orders
type OrderType string

const (
	InvalidOrder   OrderType = "INVL"
	NewOrder                 = "NWOD"
	AcceptedOrder            = "ACOD"
	CompletedOrder           = "COOD"
	InternalOrder            = "INOD"
)

type OrderMessage struct {
	Type      OrderType
	Floor     driver.Floor
	Direction driver.Direction
}

func orderToStr(order OrderMessage) string {
	str := string(order.Type) + strconv.Itoa(int(order.Floor)) + strconv.Itoa(int(order.Direction))
	crc := crc16.Crc16([]byte(str))
	// HAXHAX bitshift and convert to byte slice -> string
	str += string([]byte{byte((crc >> 8) & 0xff), byte(crc & 0xff)})
	return str
}

func strToOrder(str string) OrderMessage {
	// Check for CRC mismatch. Not tamper-proof, but should be corruption-proof.
	calculatedCrc := crc16.Crc16([]byte(str[:6]))
	receivedCrc := (uint16(str[6]) << 8) + uint16(str[7])

	if calculatedCrc != receivedCrc {
		log.Error("CRC mismatch in " + str[:4] + " message!")
		return OrderMessage{Type: InvalidOrder} // Probably corrupted
	}

	floorNum, _ := strconv.Atoi(string(str[4]))
	dirNum, _ := strconv.Atoi(string(str[4]))

	return OrderMessage{Type: OrderType(str[:4]), Floor: driver.Floor(floorNum), Direction: driver.Direction(dirNum)}

}

var udpSendCh, udpRecvCh chan udp.Udp_message

const LPORT = 13376
const BPORT = 13377
const MSGLEN = 8

func SendOrder(order OrderMessage) {
	str := orderToStr(order)
	log.Debug("Sending message: ", str)
	udpSendCh <- udp.Udp_message{Raddr: "broadcast", Data: str}
}

func InitAndHandle(receiveCh chan<- OrderMessage) {
	udpSendCh = make(chan udp.Udp_message)
	udpRecvCh = make(chan udp.Udp_message)

	udp.Udp_init(LPORT, BPORT, MSGLEN, udpSendCh, udpRecvCh)

	for {
		msg := <-udpRecvCh
		if msg.Length != 8 { // Disregard messages not 8 in length
			log.Warning("Non-8-byte message received")
			continue
		}
		order := strToOrder(msg.Data)
		receiveCh <- order
	}
}
