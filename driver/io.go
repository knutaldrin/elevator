package driver // where "driver" is the folder that contains io.go, io.c, io.h, channels.go, channels.h and driver.go
/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C"
import "time"

/*
type ButtonEvent struct {
}

func ButtonListener(chan<- ch) {
C.
}
*/

func Kek() {

	C.io_init()

	C.io_set_bit(0x300+14)
	time.Sleep(time.Second)
	C.io_clear_bit(0x300+14)
	time.Sleep(time.Second)
	C.io_set_bit(0x300+14)
	time.Sleep(time.Second)
	C.io_clear_bit(0x300+14)
}
