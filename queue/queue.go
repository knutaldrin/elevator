package main

import (
	"time"

	"github.com/knutaldrin/elevator/driver"
	"github.com/knutaldrin/elevator/log"
	"github.com/knutaldrin/elevator/net"
)

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

const t time.Duration = time.Second          //time unit
const delay time.Duration = 20 * time.Second //Delay for order accepted by other elevator

var idOffset time.Duration

var floorStatus driver.Floor
var dirStatus driver.Direction

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
func compareJob(a Job, b Job) bool {
	return (a.Direction == b.Direction && a.Floor == b.Floor)
}

//removeJob from a queue. Returns resulting queue
func removeJob(job Job, queue []Job) []Job {
	var newQueue []Job
	for i := 0; i < len(queue); i++ {
		if queue[i] != job {
			newQueue = append(newQueue, queue[i])
		}
	}
	log.Debug("Removed job: Floor %d, Direction %d", job.Floor, job.Direction)
	return newQueue
}

//moveTo moves a job from one queue to another. Returns resulting target queue.
func moveJob(job Job, from []Job, target []Job) []Job {
	removeJob(job, from)
	return append(target, job)
}

//extendTimeout of a job in a queue by t
func extendTimeout(job Job, queue []Job, t time.Duration) {
	for i := 0; i < len(queue); i++ {
		if compareJob(queue[i], job) {
			queue[i].Timeout = queue[i].Timeout.Add(t)
			return
		}
	}
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
	return 1 //TODO return int(math.Abs(float(a) - float(b))
}

//COST FUNCTION
//generateTimeout is effectively the cost function, assigning a timeout point to a job based on the status of the local elevator. A "convenient" job generates a short delay.
func generateTimeout(job Job) time.Time {
	var d time.Duration

	//diff := floorAbsDiff(job.Floor, floorStatus)

	if !isAhead(job) {
		d += t
		if dirStatus != driver.DirectionNone {
			d += t

		}

		d += idOffset
	}

	return time.Now().Add(d)
}

func generateIDOffset(id int8) time.Duration {
	return time.Duration(id) * t / 4
}

func isInQueue(job Job, queue []Job) bool {
	for i := 0; i < len(queue); i++ {
		if compareJob(queue[i], job) {
			return true
		}
	}
	return false
}

//QueueManager should be spawned as a goroutine and manages the work queues.
func QueueManager(received <-chan net.OrderMessage, id int8) {
	log.Debug("Initializing Job manager")
	idOffset = generateIDOffset(id)
	log.Debug("Generated unique offset: %d Milliseconds", idOffset.Nanoseconds()/1000000)

	for {
		select {
		case order := <-received:
			job := makeJob(order.Floor, order.Direction)
			switch order.Type {
			case net.NewOrder:
				if !isInQueue(job, myReceivedJobs) {
					job.Timeout = generateTimeout(job)
					myReceivedJobs = append(myReceivedJobs, job)
				}
				break
			case net.AcceptedOrder:
				extendTimeout(job, myReceivedJobs, delay)
				break
			case net.CompletedOrder:
				myReceivedJobs = removeJob(job, myReceivedJobs)
				break
			case net.InternalOrder:
				myActiveJobs = append(myReceivedJobs, job)
				break
			}
		}

		now := time.Now()

		for i := 0; i < len(myReceivedJobs); i++ {
			if now.After(myReceivedJobs[i].Timeout) {
				myActiveJobs = moveJob(myReceivedJobs[i], myReceivedJobs, myActiveJobs)
				log.Debug("Accepted job: Floor %d, Direction %d", myReceivedJobs[i].Floor, myReceivedJobs[i].Direction)
			}
		}
	}
}

//SetDirStatus sets the direction status
func SetDirStatus(dir driver.Direction) {
	dirStatus = dir
}

//SetFloorStatusIsTarget reports the floor and direction of the elevator to the queue, and returns whether or not it is a target (if it should stop). If yes, the relevant job is completed and removed from the queue.
func SetFloorStatusIsTarget(floor driver.Floor, dir driver.Direction) bool {
	floorStatus = floor
	if isInQueue(makeJob(floor, dir), myActiveJobs) || isInQueue(makeJob(floor, driver.DirectionNone), myActiveJobs) { //Also checks for internal orders (DirectionNone)
		return true
	}
	return false
}

//NextTarget returns the next direction that should be targeted by the elevator
func NextTarget() driver.Floor { //TODO Probably need to make more intelligent
	return myActiveJobs[0].Floor
}

func main() {

}
