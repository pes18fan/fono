package main

import "time"

type Status interface {
	isStatus()
}

type PositionUpdate struct {
	Position time.Duration
	Length   time.Duration
}

func (PositionUpdate) isStatus() {}

type PlayStateUpdate struct {
	Paused bool
}

func (PlayStateUpdate) isStatus() {}
