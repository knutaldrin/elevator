package queue

import (
	"fmt"
	"math"
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
)

//This is arbitrary and could be set to 2^32 - however this could potentially cause multiple elevators to take the same order.
const nMaxElev int = 5

//Job struct is an entry in the job queue. An expired Timeout indicates the job should be accepted and moved to myActiveJobs.
type Job struct {
	Floor     driver.Floor
	Direction driver.Direction
	Timeout   time.Time
}

/*WHAT IS AN net.OrderMessage YOU SAY?
type OrderMessage struct {
	Type      OrderType
	Floor     driver.Floor
	Direction driver.Direction
}
*/

const t time.Duration = 50 * time.Millisecond
const baseOffsetMS uint = uint(50 / nMaxElev)

const delay time.Duration = 15 * time.Second //Delay for order accepted by other elevator

var idOffset time.Duration

var floorStatus driver.Floor
var dirStatus driver.Direction = driver.DirectionNone

//Queues: Received jobs are jobs in the network. Active jobs are jobs accepted by the local elevator.
var myReceivedJobs []Job
var myActiveJobs []Job

//UTILITY FUNCTIONS
//orderToJob initializes a job with Floor and Dir from a net.OrderMessage. DOES NOT INITIALIZE TIMEOUT
func makeJob(f driver.Floor, d driver.Direction) Job {
	var j Job
	j.Floor = f
	j.Direction = d
	return j
}

//compareJob
func compareJob(a, b Job) bool {
	return (a.Direction == b.Direction && a.Floor == b.Floor)
}

func isInQueue(queue []Job, job Job) bool {
	for i := 0; i < len(queue); i++ {
		if compareJob(queue[i], job) {
			return true
		}
	}
	return false
}

//addJob to a queue. Returns resulting queue.
func addJob(queue []Job, job Job) []Job {
	if !isInQueue(queue, job) {
		queue = append(queue, job)
		log.Debug("Added job: Floor, Direction: ", job.Floor, ", ", job.Direction)
	}

	return queue
}

//removeJob from a queue. Returns resulting queue.
func removeJob(queue []Job, job Job) []Job {
	var newQueue []Job
	for i := 0; i < len(queue); i++ {
		if !compareJob(queue[i], job) {
			newQueue = append(newQueue, queue[i])
		} else {
			log.Debug("Removed job: Floor, Direction: ", job.Floor, ", ", job.Direction)
		}
	}
	return newQueue
}

//moveTo moves a job from one queue to another. Returns resulting target queue.
func moveJob(job Job, from []Job, target []Job) []Job {
	removeJob(from, job)
	return append(target, job)
}

//isAhead checks whether or not a job is ahead in the direction of travel
func isAhead(job Job) bool {
	switch dirStatus {
	case driver.DirectionUp:
		if job.Floor > floorStatus {
			return true
		}
		break
	case driver.DirectionDown:
		if job.Floor < floorStatus {
			return true
		}
		break
	}
	return false //False if no direction/idle
}

func floorAbsDiff(a, b driver.Floor) int {
	return int(math.Abs(float64(a - b)))
}

//generateIDOffset generates the unique delay for each elevator to avoid multiple elevators taking the same order.
//nMaxElev determines the "resolution" of the delay space.
func generateIDOffset(id uint) time.Duration {
	return time.Duration(id*baseOffsetMS) * time.Millisecond
}

//extendTimeout of a job in a queue by t
func extendTimeout(queue []Job, job Job, t time.Duration) {
	for i := 0; i < len(queue); i++ {
		if compareJob(queue[i], job) {
			queue[i].Timeout = queue[i].Timeout.Add(t)
			return
		}
	}
}

//COST FUNCTION
//generateTimeout is effectively the cost function, assigning a timeout point to a job based on the status of the local elevator. A "convenient" job generates a short delay.
func generateTimeout(job Job) time.Time {
	var d time.Duration

	if !isAhead(job) {
		d += t
		if dirStatus != driver.DirectionNone {
			d += t * time.Duration(floorAbsDiff(job.Floor, floorStatus)+1)
		}

		d += idOffset
		fmt.Print(d)
	}

	return time.Now().Add(d)
}

//Manager should be spawned as a goroutine and manages the work queues.
func Manager(received <-chan net.OrderMessage, id uint) {
	log.Debug("Initializing Queueueueueue manager")
	idOffset = generateIDOffset(id)
	log.Debug("Generated unique offset: ", idOffset)

	//Queue init
	myReceivedJobs = make([]Job, 0)
	myActiveJobs = make([]Job, 0)

	for {
		select {
		case order := <-received:
			job := makeJob(order.Floor, order.Direction)
			switch order.Type {
			case net.NewOrder:
				if !isInQueue(myReceivedJobs, job) {
					job.Timeout = generateTimeout(job)
					myReceivedJobs = addJob(myReceivedJobs, job)
				}
				break
			case net.AcceptedOrder:
				extendTimeout(myReceivedJobs, job, delay)
				break
			case net.CompletedOrder:
				myReceivedJobs = removeJob(myReceivedJobs, job)
				break
			case net.InternalOrder:
				myActiveJobs = addJob(myReceivedJobs, job)
				break
			}
		}

		now := time.Now()

		for i := 0; i < len(myReceivedJobs); i++ {
			if now.After(myReceivedJobs[i].Timeout) {
				myActiveJobs = moveJob(myReceivedJobs[i], myReceivedJobs, myActiveJobs)
				log.Debug("Accepted job: Floor and direction:", myReceivedJobs[i].Floor, ",", myReceivedJobs[i].Direction)
			}
		}
	}
}

//SetDirStatus sets the direction status
func SetDirStatus(dir driver.Direction) {
	dirStatus = dir
	log.Debug("Queue dir status set to", dir)
}

//ShouldStopAtFloor reports the floor and direction of the elevator to the queue, and returns whether or not the current state is a target (if it should stop).
//If yes, the relevant job is completed and removed from the queue.
func ShouldStopAtFloor(floor driver.Floor) bool {
	floorStatus = floor
	dirJob := makeJob(floor, dirStatus)
	intJob := makeJob(floor, driver.DirectionNone)
	log.Debug("Status update to queue: Floor: ", floor, " Dir: ", dirStatus)
	if isInQueue(myActiveJobs, dirJob) || isInQueue(myActiveJobs, intJob) { //Also checks for internal orders (DirectionNone)
		myActiveJobs = removeJob(myActiveJobs, intJob)
		myActiveJobs = removeJob(myActiveJobs, dirJob)
		return true
	}
	log.Debug("No jobs removed")
	log.Debug("Active queue is: ", myActiveJobs)
	return false
}

//NextDir should be spawned as a goroutine. Blocks until active queue has an entry.
func NextDir(ch chan<- driver.Direction) { //TODO Probably need to make more intelligent
	for {
		if len(myActiveJobs) != 0 {
			log.Info("Active job found")
			break
		}
	}

	if myActiveJobs[0].Floor > floorStatus {
		ch <- driver.DirectionUp
	} else if myActiveJobs[0].Floor < floorStatus {
		ch <- driver.DirectionDown
	} else {
		ch <- driver.DirectionNone
	}
}

/*func main() { //debug
	log.Warning("UTIL TEST\n")
	log.Info(floorAbsDiff(2, 2), "\n")
	log.Info(floorAbsDiff(2, 4), "\n\n")
	fmt.Print(time.Now(), "\n")

	log.Warning("QUEUE TEST\n")
	fmt.Print(NextTarget(), "\n")
	fmt.Print(makeJob(1, driver.DirectionUp), "\n")
	myActiveJobs = addJob(myActiveJobs, makeJob(1, driver.DirectionUp))
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	myActiveJobs = addJob(myActiveJobs, makeJob(3, driver.DirectionNone))
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	myActiveJobs = addJob(myActiveJobs, makeJob(55, driver.DirectionUp))
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	myActiveJobs = addJob(myActiveJobs, makeJob(1521, driver.DirectionUp))
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	myActiveJobs = addJob(myActiveJobs, makeJob(55, driver.DirectionUp))
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	myActiveJobs = removeJob(myActiveJobs, makeJob(55, driver.DirectionUp))
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	fmt.Print(ShouldStopAtFloor(1), "\n")
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	fmt.Print(ShouldStopAtFloor(1), "\n")
	fmt.Print(myActiveJobs, "\n")
	fmt.Print(NextTarget(), "\n")
	fmt.Print("\n\n")

}*/
