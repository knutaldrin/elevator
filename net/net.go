package net

import (
	"strconv"

	crc16 "github.com/joaojeronimo/go-crc16"
	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net/udp"
)

/** NEW MESSAGE FORMAT
 * Always 8 chars
 *
 * 2 chars: type
 * * NW = New order
 * * AC = Accepted order
 * * CO = Completed order
 * 1 char: Always 0, reserved for future
 * 1 char: ID
 * 1 char: floor (0-indexed)
 * 1 char: direction (0: up, 1: down)
 * 2 chars: CRC-16 of the previous 6 bytes
 */

//OrderType is an enum for communicating information about orders
type OrderType string

// Enum of order types
const (
	InvalidOrder   OrderType = "IV"
	NewOrder       OrderType = "NW"
	AcceptedOrder  OrderType = "AC"
	CompletedOrder OrderType = "CO"
)

// OrderMessage struct of a net message
type OrderMessage struct {
	Type      OrderType
	SenderID  uint
	Floor     driver.Floor
	Direction driver.Direction
}

func orderToStr(order OrderMessage) string {
	str := string(order.Type) + "0" + strconv.Itoa(int(order.SenderID)) + strconv.Itoa(int(order.Floor)) + strconv.Itoa(int(order.Direction))
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

	senderID, _ := strconv.Atoi(string(str[3]))
	floorNum, _ := strconv.Atoi(string(str[4]))
	dirNum, _ := strconv.Atoi(string(str[5]))

	return OrderMessage{Type: OrderType(str[:2]), SenderID: uint(senderID), Floor: driver.Floor(floorNum), Direction: driver.Direction(dirNum)}

}

var udpSendCh, udpRecvCh chan udp.Udp_message

// LPORT Local listen port
const LPORT = 13376

// BPORT Broadcast listen port
const BPORT = 13377

// MSGLEN Network message length
const MSGLEN = 8

var elevatorID uint

// SendOrder sends the parameter order struct to the network
func SendOrder(order OrderMessage) {
	if order.Direction == driver.DirectionNone {
		// TODO: Return instead of panicking
		log.Error("Order cannot have no direction")
		panic("OMG NO DIR")
	}
	order.SenderID = elevatorID
	str := orderToStr(order)
	log.Debug("Sending message: ", str)
	udpSendCh <- udp.Udp_message{Raddr: "broadcast", Data: str}
}

// InitAndHandle initializes network and handles receive
func InitAndHandle(receiveCh chan<- OrderMessage, id uint) {
	udpSendCh = make(chan udp.Udp_message, 8)
	udpRecvCh = make(chan udp.Udp_message, 8)

	elevatorID = id

	udp.Udp_init(LPORT, BPORT, MSGLEN, udpSendCh, udpRecvCh)

	for {
		msg := <-udpRecvCh
		if msg.Length != 8 { // Disregard messages not 8 in length
			log.Warning("Non-8-byte message received")
			continue
		}
		order := strToOrder(msg.Data)
		if order.SenderID != elevatorID { // Don't loop
			log.Info("Received order: ID: ", order.SenderID, ", type: ", order.Type, ", floor: ", order.Floor)
			receiveCh <- order
		}
	}
}
