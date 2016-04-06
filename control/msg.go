package main

type Direction int8

const (
	DOWN Direction = -1
	UP   Direction = 1
)

type Job struct {
	created   Time
	accepted  Time
	finished  Time
	floor     uint8
	direction Direction
}
