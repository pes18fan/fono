package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	tag "github.com/unitnotes/audiotag"
)

type AudioCommand int

const (
	playpause AudioCommand = iota
)

func Play(file string, statusChan chan Status, cmdChan chan AudioCommand) error {
	done := make(chan bool, 1)

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	log.Println("opened", file)

	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch {
	case strings.HasSuffix(file, ".mp3"):
		streamer, format, err = mp3.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
	case strings.HasSuffix(file, ".flac"):
		streamer, format, err = flac.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
	case strings.HasSuffix(file, ".ogg"):
		streamer, format, err = vorbis.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
	case strings.HasSuffix(file, ".wav"):
		streamer, format, err = wav.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
	default:
		return fmt.Errorf("only mp3, flac, wav and ogg formats are supported")
	}

	log.Println("set up streamer")

	go func() {
		defer streamer.Close()

		m, err := tag.ReadFrom(f)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("read tags from", file)
		statusChan <- AudioInfoUpdate{
			Artist: m.Artist(),
			Title:  m.Title(),
			Album:  m.Album(),
		}

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

		ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
		speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
			done <- true
		})))
		log.Println("playing", file)

		lastPaused := false
		for {
			select {
			case <-done:
				return
			case cmd := <-cmdChan:
				speaker.Lock()
				switch cmd {
				case playpause:
					log.Println("received playpause command")
					ctrl.Paused = !ctrl.Paused
					statusChan <- PlayStateUpdate{Paused: ctrl.Paused}
				}
				speaker.Unlock()
			case <-time.After(time.Second):
				speaker.Lock()
				statusChan <- PositionUpdate{
					Length:   format.SampleRate.D(streamer.Len()).Round(time.Second),
					Position: format.SampleRate.D(streamer.Position()).Round(time.Second),
				}
				if ctrl.Paused != lastPaused {
					statusChan <- PlayStateUpdate{Paused: ctrl.Paused}
					lastPaused = ctrl.Paused
				}
				speaker.Unlock()
			}
		}
	}()

	return nil
}
