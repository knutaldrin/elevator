package udp

import (
	"fmt"
	"net"
	"strconv"
)

var laddr *net.UDPAddr //Local address
var baddr *net.UDPAddr //Broadcast address

type Udp_message struct {
	Raddr  string //if receiving raddr=senders address, if sending raddr should be set to "broadcast" or an ip:port
	Data   string //TODO: implement another encoding, strings are meh
	Length int    //length of received data, in #bytes // N/A for sending
}

func Udp_init(localListenPort, broadcastListenPort, message_size int, send_ch, receive_ch chan Udp_message) (err error) {
	//Generating broadcast address
	baddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastListenPort))
	if err != nil {
		return err
	}

	//Generating localaddress
	tempConn, err := net.DialUDP("udp4", nil, baddr)
	defer tempConn.Close()
	tempAddr := tempConn.LocalAddr()
	laddr, err = net.ResolveUDPAddr("udp4", tempAddr.String())
	laddr.Port = localListenPort

	//Creating local listening connections
	localListenConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return err
	}

	//Creating listener on broadcast connection
	broadcastListenConn, err := net.ListenUDP("udp", baddr)
	if err != nil {
		localListenConn.Close()
		return err
	}

	go udp_receive_server(localListenConn, broadcastListenConn, message_size, receive_ch)
	go udp_transmit_server(localListenConn, broadcastListenConn, send_ch)
	return err
}

func udp_transmit_server(lconn, bconn *net.UDPConn, send_ch chan Udp_message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR in udp_transmit_server: %s \n Closing connection.\n", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	var err error
	var n int

	for {
		msg := <-send_ch
		if msg.Raddr == "broadcast" {
			n, err = lconn.WriteToUDP([]byte(msg.Data), baddr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.Raddr)
			if err != nil {
				fmt.Println("Error: udp_transmit_server: could not resolve raddr.")
				panic(err)
			}
			n, err = lconn.WriteToUDP([]byte(msg.Data), raddr)
		}
		if err != nil || n < 0 {
			fmt.Println("Error: udp_transmit_server: writing.")
			panic(err)
		}
	}
}

func udp_receive_server(lconn, bconn *net.UDPConn, message_size int, receive_ch chan Udp_message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR in udp_receive_server: %s \n Closing connection.\n", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	bconn_rcv_ch := make(chan Udp_message)
	lconn_rcv_ch := make(chan Udp_message)

	go udp_connection_reader(lconn, message_size, lconn_rcv_ch)
	go udp_connection_reader(bconn, message_size, bconn_rcv_ch)

	for {
		select {

		case buf := <-bconn_rcv_ch:
			receive_ch <- buf

		case buf := <-lconn_rcv_ch:
			receive_ch <- buf
		}
	}
}

func udp_connection_reader(conn *net.UDPConn, message_size int, rcv_ch chan Udp_message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR in udp_connection_reader: %s \n Closing connection.\n", r)
			conn.Close()
		}
	}()

	for {
		buf := make([]byte, message_size)
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil || n < 0 {
			fmt.Printf("Error: udp_connection_reader: reading\n")
			panic(err)
		}
		rcv_ch <- Udp_message{Raddr: raddr.String(), Data: string(buf), Length: n}
	}
}
