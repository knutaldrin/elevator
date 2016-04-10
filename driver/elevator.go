package driver

/*
#cgo CFLAGS: -std=c11
//#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"

type Direction C.enum_elev_motor_direction_t
type Floor uint8

type Elevator struct {
	Dir   Direction
	floor Floor
}

func Init() {
	C.elev_init()
}
