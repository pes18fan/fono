package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	termWidth  int
	termHeight int

	statusChan      chan Status
	cmdChan         chan AudioCommand
	currentPosition time.Duration
	currentLength   time.Duration
	paused          bool
}

// tea message type for status updates
type StatusMsg Status

func listenForStatus(statusChan <-chan Status) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(<-statusChan)
	}
}

func initialModel() model {
	return model{
		statusChan: make(chan Status),
		cmdChan:    make(chan AudioCommand),
	}
}

func (m model) Init() tea.Cmd {
	return listenForStatus(m.statusChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusMsg:
		switch status := msg.(type) {
		case PositionUpdate:
			m.currentPosition = status.Position
			m.currentLength = status.Length
		case PlayStateUpdate:
			m.paused = status.Paused
		}
		return m, listenForStatus(m.statusChan)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p":
			m.cmdChan <- playpause
		}
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		m.termWidth = msg.Width
	}

	return m, nil
}

func (m model) View() string {
	headingStyle := lipgloss.NewStyle().
		Width(m.termWidth).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("86"))
	s := headingStyle.Render("KASET")
	s += "\n\n"

	s += "Duration: " + m.currentLength.String() + "\n"
	s += "Position: " + m.currentPosition.String() + "\n"
	s += "Status: " + map[bool]string{true: "Paused", false: "Playing"}[m.paused] + "\n\n"
	s += "Press p to play/pause.\n"
	s += "Press q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	model := initialModel()
	p := tea.NewProgram(model)

	err := Play("homies.mp3", model.statusChan, model.cmdChan)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
