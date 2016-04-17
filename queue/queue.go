package queue

import (
	"container/list"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
)

const timeoutDelay = time.Second * 10
const delayUnit = time.Millisecond * 60

var elevID uint

var shouldStop [3][driver.NumFloors]bool

type order struct {
	floor driver.Floor
	dir   driver.Direction
	timer *time.Timer
}

var currentFloor driver.Floor
var currentDir = driver.DirectionNone

var pendingOrders = list.New()

var timeoutCh chan<- bool

// SetID sets elevator ID
func SetID(id uint) {
	elevID = id
}

// SetTimeoutCh is a channel for the queue to notify when a timer runs out, in order to wake the elvator.
func SetTimeoutCh(ch chan<- bool) {
	timeoutCh = ch
}

//ImportInternalLog imports any locally saved internal orders to the active queue. Called at init.
func ImportInternalLog() {
	intSlice := ReadLog()

	for i := 0; i < len(intSlice); i++ {
		shouldStop[driver.DirectionNone][intSlice[i]] = true
		driver.ButtonLightOn(driver.Floor(intSlice[i]), driver.DirectionNone)
	}
}

func isAhead(floor driver.Floor) bool {
	if currentDir == driver.DirectionUp {
		return floor > currentFloor
	} else if currentDir == driver.DirectionDown {
		return floor < currentFloor
	}

	return false
}

func abs(a, b int16) int16 {
	if a-b < 0 {
		return b - a
	}
	return a - b
}

func calculateTimeout(floor driver.Floor, dir driver.Direction) time.Duration {
	var delay time.Duration

	if !isAhead(floor) {
		delay += 15 * delayUnit
	}

	if dir != currentDir {
		delay += 10 * delayUnit
	}

	delay += time.Duration(abs(int16(floor), int16(currentFloor))) * delayUnit

	if currentDir == driver.DirectionNone {
		delay = delayUnit * time.Duration(elevID)
	}

	return delay
}

// Update is called when the elevator passes a floor
func Update(floor driver.Floor) {
	currentFloor = floor
}

func gotoDir(floor driver.Floor) driver.Direction {
	if floor > currentFloor {
		return driver.DirectionUp
	} else if floor < currentFloor {
		return driver.DirectionDown
	}

	return driver.DirectionNone
}

// ShouldStop at the floor?
func ShouldStop(floor driver.Floor) bool {
	if floor == 0 || floor == driver.NumFloors-1 {
		return true
	}
	return shouldStop[currentDir][floor] || shouldStop[driver.DirectionNone][floor]
}

// NextDirection gives and sets next direction
func NextDirection() driver.Direction {
	// BOOOOOOILERPLATE
	if currentDir == driver.DirectionUp {
		for i := currentFloor + 1; i < driver.NumFloors; i++ {
			if shouldStop[driver.DirectionUp][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		// then the other way
		for i := driver.NumFloors - 1; i >= 0; i-- {
			if shouldStop[driver.DirectionDown][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		for i := 0; i < int(currentFloor); i++ {
			if shouldStop[driver.DirectionUp][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
	} else {
		for i := currentFloor - 1; i >= 0; i-- {
			if shouldStop[driver.DirectionDown][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		// then the other way
		for i := 0; i < driver.NumFloors; i++ {
			if shouldStop[driver.DirectionUp][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		for i := driver.NumFloors - 1; i > int(currentFloor); i-- {
			if shouldStop[driver.DirectionDown][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
	}
	currentDir = driver.DirectionNone
	return currentDir
}

// NewOrder locally or remotely
func NewOrder(floor driver.Floor, dir driver.Direction) {
	if dir == driver.DirectionNone { // From inside the elevator
		shouldStop[dir][floor] = true
		driver.ButtonLightOn(floor, dir)
		AddToLog(int(floor)) //Log internal order to file
	} else { // From external panel on this or some other elevator

		if floor == 0 {
			dir = driver.DirectionDown
		} else if floor == driver.NumFloors-1 {
			dir = driver.DirectionUp
		}

		o := order{
			floor: floor,
			dir:   dir,
			timer: time.AfterFunc(calculateTimeout(floor, dir), func() {
				shouldStop[dir][floor] = true
				if currentDir == driver.DirectionNone {
					// Ping
					timeoutCh <- true
				}
				// Send network message that we have accepted
				net.SendOrder(net.OrderMessage{Type: net.AcceptedOrder, Floor: floor, Direction: dir})
				log.Info("Accepted order for floor ", floor)
			}),
		}

		pendingOrders.PushBack(&o)
	}
	driver.ButtonLightOn(floor, dir)
}

// OrderAcceptedRemotely yay!
func OrderAcceptedRemotely(floor driver.Floor, dir driver.Direction) {
	if floor == 0 {
		dir = driver.DirectionUp
	} else if floor == driver.NumFloors-1 {
		dir = driver.DirectionUp
	}
	// Algorithmically excellent searching
	for o := pendingOrders.Front(); o != nil; o = o.Next() {
		v := o.Value.(*order)
		if v.floor == floor && v.dir == dir {
			v.timer.Reset(timeoutDelay + calculateTimeout(floor, dir))
			return
		}
	}

	// Already completed? Maybe a late package or wtf
	log.Warning("Non-existant job accepted remotely")
}

//ClearOrderLocal is called by the local elevator, and clears both internal and external orders. Calls ClearOrder.
func ClearOrderLocal(floor driver.Floor, dir driver.Direction) {
	// Turn off inside too
	shouldStop[driver.DirectionNone][floor] = false
	driver.ButtonLightOff(floor, driver.DirectionNone)
	dir = currentDir
	RemoveFromLog(int(floor))
	ClearOrder(floor, dir)
}

// ClearOrder means an order is completed (either remotely or locally). Does not clear internal orders, but is called by ClearOrderLocal.
func ClearOrder(floor driver.Floor, dir driver.Direction) {
	if floor == 0 {
		dir = driver.DirectionDown
	} else if floor == driver.NumFloors-1 {
		dir = driver.DirectionUp
	}
	shouldStop[dir][floor] = false
	driver.ButtonLightOff(floor, dir)

	// Clear from pendingOrders
	for o := pendingOrders.Front(); o != nil; o = o.Next() {
		v := o.Value.(*order)
		if v.floor == floor && v.dir == dir {
			v.timer.Stop()
			pendingOrders.Remove(o)
			break
		}
	}
}
