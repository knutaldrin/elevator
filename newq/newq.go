package newq

import (
	"container/list"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
)

// TODO TODO TODO: There is no sensible reason why lights should be controlled here

const timeoutDelay = time.Second * 20

var shouldStop [3][driver.NumFloors]bool

type order struct {
	floor driver.Floor
	dir   driver.Direction
	timer *time.Timer
}

var currentFloor driver.Floor
var currentDir driver.Direction

var pendingOrders = list.New()

var timeoutCh chan<- bool

func SetTimeoutCh(ch chan<- bool) {
	timeoutCh = ch
}

func calculateTimeout(floor driver.Floor, dir driver.Direction) time.Duration {
	return time.Second // TODO: Be sensible
}

func getOrder(floor driver.Floor, dir driver.Direction) *list.Element {
	for o := pendingOrders.Front(); o != nil; o.Next() {
		return o
	}
	return nil
}

func Update(floor driver.Floor) {
	currentFloor = floor
}

func ShouldStop(floor driver.Floor) bool {
	reverse := true

	for i := 0; i < driver.NumFloors; i++ {
		if shouldStop[currentDir][floor] == true {
			reverse = false
			break
		}
	}

	if reverse {
		if currentDir == driver.DirectionDown {
			return shouldStop[driver.DirectionUp][floor]
		} else {
			return shouldStop[driver.DirectionDown][floor]
		}
	}

	return shouldStop[currentDir][floor] || shouldStop[driver.DirectionNone][floor] //||
	//floor == 0 || floor == driver.NumFloors-1
}

// NextDirection gives and sets next direction
func NextDirection() driver.Direction {
	// BOOOOOOILERPLATE
	if currentDir == driver.DirectionUp {
		for i := currentFloor + 1; i < driver.NumFloors; i++ {
			if shouldStop[driver.DirectionUp][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = driver.DirectionUp
				return currentDir
			}
		}
		// then the other way
		for i := driver.NumFloors - 1; i >= 0; i-- {
			if shouldStop[driver.DirectionDown][i] || shouldStop[driver.DirectionNone][i] {
				if driver.Floor(i) < currentFloor {
					currentDir = driver.DirectionDown
				} else {
					currentDir = driver.DirectionUp
				}
				return currentDir
			}
		}
	} else {
		for i := currentFloor - 1; i >= 0; i-- {
			if shouldStop[driver.DirectionDown][i] || shouldStop[driver.DirectionNone][i] {
				currentDir = driver.DirectionDown
				return currentDir
			}
		}
		// then the other way
		for i := 0; i < driver.NumFloors; i++ {
			if shouldStop[driver.DirectionUp][i] || shouldStop[driver.DirectionNone][i] {
				if driver.Floor(i) < currentFloor {
					currentDir = driver.DirectionDown
				} else {
					currentDir = driver.DirectionUp
				}
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
		//shouldStop[driver.DirectionUp][floor] = true
		//shouldStop[driver.DirectionDown][floor] = true
		shouldStop[driver.DirectionNone][floor] = true

		// TODO: Save to log
	} else { // From external panel on this or some other elevator

		o := order{
			floor: floor,
			dir:   dir,
			timer: time.AfterFunc(calculateTimeout(floor, dir), func() {
				/*if floor == 0 {
					shouldStop[driver.DirectionUp][floor] = true
				} else if floor == driver.NumFloors-1 {
					shouldStop[driver.DirectionDown][floor] = true
				} else {*/
				shouldStop[dir][floor] = true
				//}
				if currentDir == driver.DirectionNone {
					// Ping
					timeoutCh <- true
				}
				// TODO: Send network message that we have accepted
			}),
		}

		pendingOrders.PushBack(&o)
		driver.ButtonLightOn(floor, dir)
	}
}

// OrderAcceptedRemotely yay!
func OrderAcceptedRemotely(floor driver.Floor, dir driver.Direction) {
	// Algorithmically excellent searching
	/*
		for o := pendingOrders.Front(); o != nil; o.Next() {
			v := o.Value.(*order)
			if v.floor == floor && v.dir == dir {
	*/
	o := getOrder(floor, dir).Value.(*order)
	if o != nil {
		o.timer.Reset(timeoutDelay + calculateTimeout(floor, dir))
	}
	//}

	// Already completed? Maybe a late package or wtf
	log.Warning("Non-existant job accepted remotely")
}

// ClearOrder means an order is completed (either remotely or locally)
func ClearOrder(floor driver.Floor, dir driver.Direction) {
	// Pop from queue
	e := getOrder(floor, dir)
	if e != nil {
		e.Value.(*order).timer.Stop()
		pendingOrders.Remove(e)
	}

	// Will never happen
	if dir == driver.DirectionNone {
		// Clear both
		// TODO: Don't, just clear the one in direction of travel
		shouldStop[currentDir][floor] = false
		driver.ButtonLightOff(floor, currentDir)
	} else {
		shouldStop[dir][floor] = false
		shouldStop[driver.DirectionNone][floor] = false
		driver.ButtonLightOff(floor, dir)
	}
	// Turn off inside too
	driver.ButtonLightOff(floor, driver.DirectionNone)
}
