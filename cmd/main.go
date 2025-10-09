package main

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type activeScreen int

const (
	nowPlayingScreen activeScreen = iota
	songSelectScreen
)

// HACK: Currently the filepicker and now playing model are both stuffed in
// one model struct and chosen based on an enum value. It works, but is not
// really an ideal way to do it; so you should probably change this to be
// based on a multi-model structure, as bubbletea does support that sort of
// UI hierarchy.
type model struct {
	activeScreen activeScreen
	quitting     bool

	termWidth  int
	termHeight int

	fileChan   chan string
	statusChan chan Status
	cmdChan    chan AudioCommand

	currentPosition time.Duration
	currentLength   time.Duration
	playState       PlayState
	currentArtist   string
	currentTitle    string
	currentAlbum    string

	progress progress.Model

	filepicker   filepicker.Model
	selectedFile string

	err error
}

// tea message type for status updates
type statusMsg Status
type clearErrorMsg struct{}

func listenForStatus(statusChan <-chan Status) tea.Cmd {
	return func() tea.Msg {
		return statusMsg(<-statusChan)
	}
}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func initialModel() model {
	prog := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	fp := filepicker.New()
	fp.AllowedTypes = []string{".mp3", ".flac", ".wav", ".ogg"}
	fp.DirAllowed = true

	return model{
		activeScreen: songSelectScreen,
		quitting:     false,

		fileChan:   make(chan string),
		statusChan: make(chan Status),
		cmdChan:    make(chan AudioCommand),

		currentPosition: 0,
		currentLength:   0,
		playState:       noTrackLoaded,
		currentArtist:   "",
		currentTitle:    "",
		currentAlbum:    "",

		progress: prog,

		filepicker:   fp,
		selectedFile: "",

		err: nil,
	}
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		switch status := msg.(type) {
		case PositionUpdate:
			m.currentPosition = status.Position
			m.currentLength = status.Length
			log.Println("playback position updated")
		case PlayStateUpdate:
			m.playState = status.PlayState
			log.Println("playback state updated")
		case AudioInfoUpdate:
			m.currentArtist = status.Artist
			m.currentTitle = status.Title
			m.currentAlbum = status.Album
			log.Println("audio info updated")
		case ErrorUpdate:
			log.Fatalf("audio playback failed: %v", status.Err)
		}
		return m, listenForStatus(m.statusChan)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "p":
			m.cmdChan <- playpause
		case "f":
			if m.activeScreen == nowPlayingScreen {
				m.activeScreen = songSelectScreen
				m.selectedFile = ""
				m.cmdChan <- stop
				return m, m.filepicker.Init()
			}
		}
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		m.termWidth = msg.Width
	case clearErrorMsg:
		m.err = nil
	}

	if m.activeScreen == songSelectScreen {
		var cmd tea.Cmd
		m.filepicker, cmd = m.filepicker.Update(msg)

		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.activeScreen = nowPlayingScreen
			m.selectedFile = path
			m.currentTitle = path // Will be overriden if valid tags are found
			m.fileChan <- path
			return m, listenForStatus(m.statusChan)
		}

		if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
			m.err = errors.New(path + " is not valid.")
			m.selectedFile = ""
			return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
		}

		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	centeredStyle := lipgloss.NewStyle().
		Width(m.termWidth).
		Align(lipgloss.Center)
	centered := centeredStyle.Render

	s := "\n"
	s += centeredStyle.
		Foreground(lipgloss.Color("#4fefca")).
		Bold(true).
		Render("Fono")
	s += "\n\n"

	if m.activeScreen == nowPlayingScreen {
		tags := centeredStyle.Foreground(lipgloss.Color("#4fefb0")).Render

		s += tags(m.currentTitle)
		if m.currentTitle != "" {
			s += "\n"
		}

		s += tags(m.currentArtist)
		if m.currentArtist != "" {
			s += "\n"
		}

		s += tags(m.currentAlbum)
		if m.currentAlbum != "" {
			s += "\n"
		}

		s += "\n\n"

		s += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4fefca")).
			Render("Status: ")
		s += map[PlayState]string{
			paused:        "Paused",
			playing:       "Playing",
			noTrackLoaded: "Track not loaded",
		}[m.playState]
		s += "\n\n"

		percent := 0.0
		if m.currentLength != 0 {
			percent = float64(m.currentPosition) / float64(m.currentLength)
		}

		// TODO: Make the progress bar animated, following the docs by Charm
		s += centeredStyle.
			Bold(true).
			Foreground(lipgloss.Color("#4fefca")).
			Render("Progress")
		s += "\n"
		s += centered(m.progress.ViewAs(percent))
		s += "\n\n"

		s += "\n"
		s += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4fefca")).
			Render("Controls: ")
		s += "\n"
		s += "    Press p to play/pause.\n"
		s += "    Press f to pick another file.\n"
		s += "    Press q to quit.\n"
	} else {
		if m.err != nil {
			s += centered(
				m.filepicker.Styles.DisabledFile.Render(m.err.Error()),
			)
		} else if m.selectedFile == "" {
			s += centered("Pick a file:")
		} else {
			s += centered(
				"Selected file: " +
					m.filepicker.Styles.Selected.Render(m.selectedFile),
			)
		}
		s += "\n\n" + m.filepicker.View() + "\n"
	}

	return s
}

func main() {
	// Set up a logger to log to `debug.log` if the `DEBUG` environment variable
	// has a value.
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			log.Fatal("failed to set up logger:", err)
		}
		defer f.Close()
	} else {
		log.SetOutput(io.Discard)
	}

	model := initialModel()
	p := tea.NewProgram(model)
	log.Println("set up tea program")

	go StartPlayer(model.fileChan, model.statusChan, model.cmdChan)
	if _, err := p.Run(); err != nil {
		log.Fatalf("tea program got error: %v", err)
	}
}
