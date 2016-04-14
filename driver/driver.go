package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
#include "io.h"
*/
import "C"
import "github.com/knutaldrin/elevator/log"

// NumFloors = number of floors in elevator
const NumFloors = 4

// Direction of travel
type Direction int8

// enum definitions for direction
const (
	DirectionUp   Direction = 0
	DirectionDown           = 1
	DirectionNone           = 2
)

// Floor is a floor. negative -> invalid
type Floor int16

// ButtonEvent for use in button listener
type ButtonEvent struct {
	Dir   Direction
	Floor Floor
}

func setFloorIndicator(floor Floor) {
	C.elev_set_floor_indicator(C.int(floor))
}

func getFloor() Floor {
	return Floor(C.elev_get_floor_sensor_signal())
}

// Init initializes the elevator, resets all lamps.
func Init() {
	C.elev_init()
}

// Reset makes sure the elevator is at a safe floor on startup
// Blocking, should never be called when listeners are running
func Reset() {
	currentFloor := getFloor()

	if currentFloor == -1 {
		log.Warning("Unknown floor")
		// Move down until we hit something
		RunDown()
		for {
			currentFloor = getFloor()
			if currentFloor != -1 {
				log.Info("At floor ", currentFloor, ", ready for service")
				setFloorIndicator(currentFloor)
				break
			}
		}
		Stop()
		// TODO: Open door?
	}
}

// OpenDoor opens the door
func OpenDoor() {
	C.elev_set_door_open_lamp(1)
}

// CloseDoor closes the door
func CloseDoor() {
	C.elev_set_door_open_lamp(0)
}

// ButtonLightOn turns on the corresponding lamp
func ButtonLightOn(floor Floor, dir Direction) {
	C.elev_set_button_lamp(C.elev_button_type_t(dir), C.int(floor), 1)
}

// ButtonLightOff turns it off
func ButtonLightOff(floor Floor, dir Direction) {
	C.elev_set_button_lamp(C.elev_button_type_t(dir), C.int(floor), 0)
}

// RunUp runs up
func RunUp() {
	if getFloor() == NumFloors-1 {
		log.Error("Trying to go up from the top floor?!")
		return
	}
	//queue.setDirStatus(1)
	C.elev_set_motor_direction(1)
}

// RunDown runs down
func RunDown() {
	if getFloor() == 0 {
		log.Error("Trying to go down from the bottom floor?!")
		return
	}
	C.elev_set_motor_direction(-1)
}

// Stop stops the elevator
func Stop() {
	C.elev_set_motor_direction(0)
}

// FloorListener sends event on floor update
func FloorListener(ch chan<- Floor) {
	currentFloor := getFloor()
	for {
		newFloor := getFloor()
		if newFloor > -1 {
			if newFloor != currentFloor {
				currentFloor = newFloor
				setFloorIndicator(newFloor)
				log.Info("Now at floor ", newFloor)
				ch <- newFloor
			}
		}
	}
}

// StopButtonListener should be spawned as a goroutine, and will trigger on press
func StopButtonListener(ch chan<- bool) {
	var stopButtonState bool

	for {
		newState := C.elev_get_stop_signal() != 0
		if newState != stopButtonState {
			stopButtonState = newState

			if newState {
				log.Debug("Stop button pressed")
			} else {
				log.Debug("Stop button released")
			}
			ch <- newState
		}
	}
}

// FloorButtonListener should be spawned as a goroutine
func FloorButtonListener(ch chan<- ButtonEvent) {
	var floorButtonState [3][NumFloors]bool

	for {
		for direction := DirectionUp; direction <= DirectionNone; direction++ {
			for floor := Floor(0); floor < NumFloors; floor++ {
				newState := C.elev_get_button_signal(C.elev_button_type_t(direction), C.int(floor)) != 0
				if newState != floorButtonState[direction][floor] {
					floorButtonState[direction][floor] = newState

					// Only dispatch an event if it's pressed
					if newState {
						log.Debug("Button type ", direction, " floor ", floor, " pressed")
						ch <- ButtonEvent{Dir: direction, Floor: floor}
					} else {
						log.Bullshit("Button type ", direction, " floor ", floor, " released")
					}
				}
			}
		}
	}
}

// TODO: Should this be a separate one?
// CommandButtonListener listens for... command buttons
/*
func CommandButtonListener(ch chan<- Floor) {
	var commandButtonState [NumFloors]bool

	for {
		for floor := Floor(0); floor < NumFloors; floor++ {
			newState := C.elev_get_button_signal(C.elev_button_type_t(DirectionNone), C.int(floor)) != 0
			if newState != commandButtonState[floor] {
				commandButtonState[floor] = newState

				// Only dispatch an event if it's pressed
				if newState {
					log.Debug("Command buttonf for floor ", floor, " pressed")
					ch <- floor
				} else {
					log.Bullshit("Command button for floor ", floor, " released")
				}
			}
		}
	}
}
*/
