package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
)

type AudioCommand int

const (
	playpause AudioCommand = iota
	stop
)

const POSITION_UPDATE_INTERVAL = time.Second
const RESAMPLE_QUALITY_FACTOR = 4

// Start up the audio unit.
// The unit accepts, processes and plays audio files over a channel.
// It should be run in a goroutine separate from the main one.
func StartPlayer(
	fileChan <-chan string,
	statusChan chan<- Status,
	cmdChan <-chan AudioCommand,
) {
	initialized := false
	speakerSampleRate := beep.SampleRate(-1)

outer:
	for {
		log.Println("waiting to get file from fileChan...")
		file, ok := <-fileChan
		if !ok {
			statusChan <- ErrorUpdate{
				fmt.Errorf("file channel unexpectedly closed"),
			}
			continue outer
		}
		log.Println("got file", file, "from fileChan")

		f, err := os.Open(file)
		if err != nil {
			statusChan <- ErrorUpdate{
				fmt.Errorf("failed to open file: %w", err),
			}
			continue outer
		}
		// Closing the streamer later will close the file itself, so don't defer close it here
		log.Println("opened", file)

		// Grab metadata
		var artist, title, album string
		var art terminalImage
		artBytes := []byte{}
		m, err := tag.ReadFrom(f)
		if err != nil {
			log.Println("failed to read tags from", file, ":", err)
		} else {
			artist = m.Artist()
			title = m.Title()
			album = m.Album()

			pic := m.Picture()
			if pic != nil {
				artBytes = pic.Data
			} else {
				log.Println("no artwork found")
			}

			log.Println("read tags from", file)
		}

		art, err = getEncodedImage(artBytes)
		if err != nil {
			log.Println("failed to read artwork:", err)
		}

		// Seek the file back to the start before creating the streamer
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			f.Close()
			statusChan <- ErrorUpdate{
				fmt.Errorf("failed to seek back to start of file: %w", err),
			}
			continue outer
		}

		var streamer beep.StreamSeekCloser
		var format beep.Format

		// Create a streamer for the file if its in a supported format
		switch {
		case strings.HasSuffix(file, ".mp3"):
			streamer, format, err = mp3.Decode(f)
		case strings.HasSuffix(file, ".flac"):
			streamer, format, err = flac.Decode(f)
		case strings.HasSuffix(file, ".ogg"):
			streamer, format, err = vorbis.Decode(f)
		case strings.HasSuffix(file, ".wav"):
			streamer, format, err = wav.Decode(f)
		default:
			err = fmt.Errorf("only mp3, flac, wav and ogg formats are supported")
		}
		if err != nil {
			f.Close()
			statusChan <- ErrorUpdate{
				fmt.Errorf("failed to decode audio file: %w", err),
			}
			continue outer
		}

		log.Println("set up streamer")

		// Careful not to double-initialize the speaker!
		if !initialized {
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			speakerSampleRate = format.SampleRate
			initialized = true
		}

		// If the chosen file has a different sample rate than that of the
		// initialized speaker, we need to resample it to make it sound right
		var resampled *beep.Resampler
		resampled = nil
		if speakerSampleRate != format.SampleRate && speakerSampleRate != -1 {
			resampled = beep.Resample(
				RESAMPLE_QUALITY_FACTOR,
				format.SampleRate,
				speakerSampleRate,
				streamer,
			)
		}

		// Send the metadata out, only if the file had at least a title
		// Otherwise we'd have no info showing on the player, we want at least
		// the filename to show up
		if title != "" {
			statusChan <- AudioInfoUpdate{
				Artist: artist,
				Title:  title,
				Album:  album,
				Art:    art,
			}
		}

		var ctrl *beep.Ctrl
		if resampled != nil {
			ctrl = &beep.Ctrl{Streamer: resampled, Paused: false}
		} else {
			ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
		}

		done := make(chan bool)
		speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
			done <- true
		})))
		log.Println("playing", file)

		statusChan <- PlayStateUpdate{
			PlayState: playing,
		}

		ticker := time.NewTicker(POSITION_UPDATE_INTERVAL)

		lastPaused := false
		for {
			select {
			case <-done:
				log.Println("finished playing track", file)
				// don't lock the speaker before clearing
				// this is cuz speaker.Clear() already tries to lock it
				speaker.Clear()

				speaker.Lock()
				streamer.Close()
				speaker.Unlock()

				statusChan <- PlayStateUpdate{PlayState: noTrackLoaded}
				statusChan <- PositionUpdate{Length: 0, Position: 0}
				statusChan <- AudioInfoUpdate{
					Artist: "",
					Title:  "",
					Album:  "",
					Art:    terminalImage{},
				}
				ticker.Stop()
				continue outer
			case cmd := <-cmdChan:
				switch cmd {
				case playpause:
					speaker.Lock()
					log.Println("received playpause command")
					ctrl.Paused = !ctrl.Paused

					var state PlayState
					if ctrl.Paused {
						state = paused
					} else {
						state = playing
					}
					speaker.Unlock()

					statusChan <- PlayStateUpdate{PlayState: state}
				case stop:
					log.Println("received stop command")
					speaker.Clear()

					speaker.Lock()
					streamer.Close()
					speaker.Unlock()

					statusChan <- PlayStateUpdate{PlayState: noTrackLoaded}
					statusChan <- PositionUpdate{Length: 0, Position: 0}
					statusChan <- AudioInfoUpdate{
						Artist: "",
						Title:  "",
						Album:  "",
						Art:    terminalImage{},
					}
					ticker.Stop()
					continue outer
				}
			case <-ticker.C:
				speaker.Lock()
				// don't bother sending position updates if paused
				if !ctrl.Paused {
					statusChan <- PositionUpdate{
						Length:   format.SampleRate.D(streamer.Len()).Round(time.Second),
						Position: format.SampleRate.D(streamer.Position()).Round(time.Second),
					}
				}

				if ctrl.Paused != lastPaused {
					lastPaused = ctrl.Paused

					var state PlayState
					if ctrl.Paused {
						state = paused
					} else {
						state = playing
					}

					statusChan <- PlayStateUpdate{PlayState: state}
				}

				if streamer.Err() != nil {
					streamer.Close()
					statusChan <- PlayStateUpdate{PlayState: noTrackLoaded}
					speaker.Unlock()
					ticker.Stop()
					continue outer
				}
				speaker.Unlock()
			}
		}
	}
}
