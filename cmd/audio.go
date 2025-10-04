package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type AudioCommand int

const (
	playpause AudioCommand = iota
)

// status updated on every second of playback
type AudioStatus struct {
	length   time.Duration
	position time.Duration
}

func Play(file string, status chan AudioStatus, commander chan AudioCommand) error {
	done := make(chan bool, 1)

	if !strings.HasSuffix(file, ".mp3") {
		return fmt.Errorf("only the mp3 format is supported")
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer streamer.Close()

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

		ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
		speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
			done <- true
		})))

		for {
			select {
			case <-done:
				return
			case <-time.After(time.Second):
				speaker.Lock()
				status <- AudioStatus{
					length:   format.SampleRate.D(streamer.Len()).Round(time.Second),
					position: format.SampleRate.D(streamer.Position()).Round(time.Second),
				}
				speaker.Unlock()
			}
		}
	}()

	return nil
}
