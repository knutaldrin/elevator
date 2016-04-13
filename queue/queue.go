package queue

import (
	"time"

	"github.com/knutaldrin/elevator/driver"
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

const t time.Duration = time.Second      //time unit
const d time.Duration = 20 * time.Second //Delay for order accepted by other elevator

var floorStatus driver.Floor
var dirStatus driver.Direction

//Queues: Received jobs are jobs in the network. Active jobs are jobs accepted by the local elevator.
var myReceivedJobs []Job
var myActiveJobs []Job

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

func removeJob(job Job, queue []Job) {
	var newQueue []Job
	for i := 0; i < len(queue); i++ {
		if queue[i] != job {
			newQueue = append(newQueue, queue[i])
		}
	}
	queue = newQueue
}

//moveTo moves a job from one queue to another.
func moveJob(job Job, from []Job, target []Job) {
	target = append(target, job)
	removeJob(job, from)
}

//extendTimeout of a job in a queue
func extendTimeout(job Job, queue []Job, t time.Duration) {
	for i := 0; i < len(queue); i++ {
		if compareJob(queue[i], job) {
			queue[i].Timeout = queue[i].Timeout.Add(t)
			return
		}
	}
}

//generateTimeout is effectively the cost function, assigning a timeout point to a job based on the status of the local elevator. A "convenient" job generates a short delay.
func generateTimeout(floorStatus driver.Floor, dirStatus driver.Direction, job Job) time.Time {
	//TODO
	return (time.Now()).Add(10 * time.Second)
}

func isInQueue(job Job, queue []Job) bool {
	for i := 0; i < len(queue); i++ {
		if compareJob(queue[i], job) {
			return true
		}
	}
	return false
}

//JobManager should be spawned as a goroutine and manages the work queues.
func JobManager(receive <-chan net.OrderMessage) {
	for {
		select {
		case order := <-receive:
			job := makeJob(order.Floor, order.Direction)
			switch order.Type {
			case net.NewOrder:
				if !isInQueue(job, myReceivedJobs) {
					job.Timeout = generateTimeout(floorStatus, dirStatus, job)
					myReceivedJobs = append(myReceivedJobs, job)
				}
				break
			case net.AcceptedOrder:
				extendTimeout(job, myReceivedJobs, d)
				break
			case net.CompletedOrder:
				removeJob(job, myReceivedJobs)
				break
			}
		}

		now := time.Now()

		for i := 0; i < len(myReceivedJobs); i++ {
			if now.After(myReceivedJobs[i].Timeout) {
				moveJob(myReceivedJobs[i], myReceivedJobs, myActiveJobs)
			}
		}
	}
}

//NextDirection returns the next direction that should be targeted by the elevator
func NextDirection() driver.Direction { //TODO Probably need to make more intelligent
	if myActiveJobs[0].Floor > floorStatus {
		return driver.DirectionUp
	}
	return driver.DirectionDown
}

//isJobTarget reports the status of the elevator to the queue, and returns whether or not it is a target (if it should stop). If yes, the relevant job is completed and removed from the queue.
func isJobTarget(floor driver.Floor, dir driver.Direction) bool {
	if isInQueue(makeJob(floor, dir), myActiveJobs) {
		return true
	}
	return false
}
