package newq

import (
	"container/list"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
)

// TODO TODO TODO: There is no sensible reason why lights should be controlled here

const timeoutDelay = time.Second * 10

var shouldStop [2][driver.NumFloors]bool

type order struct {
	floor driver.Floor
	dir   driver.Direction
	timer *time.Timer
}

var currentFloor driver.Floor
var currentDir = driver.DirectionNone

var pendingOrders = list.New()

var timeoutCh chan<- bool

// SetTimeoutCh is a channel for the queue to notify when a timer runs out, in order to wake the elvator.
func SetTimeoutCh(ch chan<- bool) {
	timeoutCh = ch
}

func calculateTimeout(floor driver.Floor, dir driver.Direction) time.Duration {
	if dir == driver.DirectionDown {
		return time.Second
	}
	return time.Second * 2
	// TODO: Be sensible
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
	return shouldStop[currentDir][floor]
}

// NextDirection gives and sets next direction
func NextDirection() driver.Direction {
	// BOOOOOOILERPLATE
	if currentDir == driver.DirectionUp {
		for i := currentFloor + 1; i < driver.NumFloors; i++ {
			if shouldStop[driver.DirectionUp][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		// then the other way
		for i := driver.NumFloors - 1; i >= 0; i-- {
			if shouldStop[driver.DirectionDown][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		for i := 0; i < int(currentFloor); i++ {
			if shouldStop[driver.DirectionUp][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
	} else {
		for i := currentFloor - 1; i >= 0; i-- {
			if shouldStop[driver.DirectionDown][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		// then the other way
		for i := 0; i < driver.NumFloors; i++ {
			if shouldStop[driver.DirectionUp][i] {
				currentDir = gotoDir(driver.Floor(i))
				return currentDir
			}
		}
		for i := driver.NumFloors - 1; i > int(currentFloor); i-- {
			if shouldStop[driver.DirectionDown][i] {
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
		shouldStop[driver.DirectionUp][floor] = true
		shouldStop[driver.DirectionDown][floor] = true

		// TODO: Save to log
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
				// TODO: Send network message that we have accepted
				net.SendOrder(net.OrderMessage{Type: net.AcceptedOrder, Floor: floor, Direction: dir})
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

// ClearOrder means an order is completed (either remotely or locally)
func ClearOrder(floor driver.Floor, dir driver.Direction) {

	// TODO: Is this needed??? Probably not, remove plx
	if dir == driver.DirectionNone {
		shouldStop[currentDir][floor] = false
	} else {
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
	// Turn off inside too
	driver.ButtonLightOff(floor, driver.DirectionNone)
}
