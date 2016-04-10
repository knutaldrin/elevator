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

// Direction of travel: -1 = down, 0 = stop, 1 = up
type Direction int8

// Floor is kinda self-explanatory. Duh.
type Floor int8

// ButtonType differs between up, down and internal buttons
type ButtonType uint8

// The floor button types
const (
	ButtonUp = ButtonType(iota)
	ButtonDown
	ButtonCommand
)

// ButtonEvent for use in button listener
type ButtonEvent struct {
	kind  ButtonType
	floor Floor
}

var currentFloor Floor = -1 // -1 is unknown

// Init initializes the elevator, resets all lamps.
func Init() {
	C.elev_init()
}

// Reset makes sure the elevator is at a safe floor on startup
// Blocking, should never be called when listeners are running
func Reset() {
	currentFloor = Floor(C.elev_get_floor_sensor_signal())

	if currentFloor == -1 {
		log.Warning("Unknown floor")
		// Move down until we hit something
		RunDown()
		for {
			currentFloor = Floor(C.elev_get_floor_sensor_signal())
			if currentFloor != -1 {
				log.Info("Now at floor ", currentFloor)
				setFloorIndicator(currentFloor)
				break
			}
		}
		Stop()
		// TODO: Open door?
	}
}

// GetFloor returns the current floor. -1 is unknown.
// TODO: Is this needed?
func GetFloor() Floor {
	return currentFloor
}

// OpenDoor opens the door
func OpenDoor() {
	C.elev_set_door_open_lamp(1)
}

// CloseDoor closes the door
func CloseDoor() {
	C.elev_set_door_open_lamp(0)
}

// RunUp runs up
func RunUp() {
	C.elev_set_motor_direction(1)
}

// RunDown runs down
func RunDown() {
	C.elev_set_motor_direction(-1)
}

// Stop stops the elevator
func Stop() {
	C.elev_set_motor_direction(0)
}

func setFloorIndicator(floor Floor) {
	C.elev_set_floor_indicator(C.int(floor))
}

// FloorListener sends event on floor update
func FloorListener(ch chan<- Floor) {
	for {
		newFloor := Floor(C.elev_get_floor_sensor_signal())
		if newFloor > -1 {
			if newFloor != currentFloor {
				currentFloor = newFloor
				setFloorIndicator(newFloor)
				log.Debug("New floor: ", newFloor)
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
		for buttonType := ButtonUp; buttonType <= ButtonCommand; buttonType++ {
			for floor := Floor(0); floor < NumFloors; floor++ {
				newState := C.elev_get_button_signal(C.elev_button_type_t(buttonType), C.int(floor)) != 0
				if newState != floorButtonState[buttonType][floor] {
					floorButtonState[buttonType][floor] = newState

					// Only dispatch an event if it's pressed
					if newState {
						log.Debug("Button type ", buttonType, " floor ", floor, " pressed")
						ch <- ButtonEvent{kind: buttonType, floor: floor}
					} else {
						log.Bullshit("Button type ", buttonType, " floor ", floor, " released")
					}
				}
			}
		}
	}
}
