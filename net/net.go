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

/** NEW MESSAGE FORMAT
 * Always 8 chars
 *
 * 2 chars: type
 * * NW = New order
 * * AC = Accepted order
 * * CO = Completed order
 * 1 char: Always 0
 * 1 char: ID
 * 1 char: floor (0-indexed)
 * 1 char: direction (0: up, 1: down)
 * 2 chars: CRC-16 of the previous 6 bytes
 */

//OrderType is an enum for communicating information about orders
type OrderType string

const (
	InvalidOrder   OrderType = "IV"
	NewOrder                 = "NW"
	AcceptedOrder            = "AC"
	CompletedOrder           = "CO"
)

type OrderMessage struct {
	Type      OrderType
	SenderId  uint
	Floor     driver.Floor
	Direction driver.Direction
}

func orderToStr(order OrderMessage) string {
	str := string(order.Type) + "0" + strconv.Itoa(int(order.SenderId)) + strconv.Itoa(int(order.Floor)) + strconv.Itoa(int(order.Direction))
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

	senderId, _ := strconv.Atoi(string(str[3]))
	floorNum, _ := strconv.Atoi(string(str[4]))
	dirNum, _ := strconv.Atoi(string(str[5]))

	return OrderMessage{Type: OrderType(str[:2]), SenderId: uint(senderId), Floor: driver.Floor(floorNum), Direction: driver.Direction(dirNum)}

}

var udpSendCh, udpRecvCh chan udp.Udp_message

const LPORT = 13376
const BPORT = 13377
const MSGLEN = 8

var elevatorId uint

func SendOrder(order OrderMessage) {
	order.SenderId = elevatorId
	str := orderToStr(order)
	log.Debug("Sending message: ", str)
	udpSendCh <- udp.Udp_message{Raddr: "broadcast", Data: str}
}

func InitAndHandle(receiveCh chan<- OrderMessage, id uint) {
	udpSendCh = make(chan udp.Udp_message, 8)
	udpRecvCh = make(chan udp.Udp_message, 8)

	elevatorId = id

	udp.Udp_init(LPORT, BPORT, MSGLEN, udpSendCh, udpRecvCh)

	for {
		msg := <-udpRecvCh
		if msg.Length != 8 { // Disregard messages not 8 in length
			log.Warning("Non-8-byte message received")
			continue
		}
		order := strToOrder(msg.Data)
		if order.SenderId != elevatorId { // Don't loop
			log.Info("Received order: ID: ", order.SenderId, ", type: ", order.Type, ", floor: ", order.Floor)
			receiveCh <- order
		}
	}
}
