package main

import "time"

// A status message sent out by the beep audio unit to the bubbletea UI.
// Status structs, alongside acting as a notifier for changes, also provide
// information about the change.
type Status interface {
	isStatus()
}

// Status update sent to signify a change in playing position of the track.
// Sent out every second by default.
type PositionUpdate struct {
	Position time.Duration
	Length   time.Duration
}

func (PositionUpdate) isStatus() {}

// Status update sent out when a track is paused or unpaused.
type PlayStateUpdate struct {
	Paused bool
}

func (PlayStateUpdate) isStatus() {}

// Status update sent when the metadata of a track changes.
// This is typically sent out when the playing track changes.
type AudioInfoUpdate struct {
	Artist string
	Title  string
	Album  string
}

func (AudioInfoUpdate) isStatus() {}
