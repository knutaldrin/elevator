package queue

import "time"

type JobStatus int8

const (
	InvalidOrder   JobStatus = 0
	NewOrder                 = 1
	AcceptedOrder            = 2
	CompletedOrder           = 3
)

type Job struct {
	floor   int
	status  JobStatus
	timeout time.Time
}

var MyJobs []Job
var BackLogJobs []Job
