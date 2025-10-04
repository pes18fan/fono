package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	status    chan AudioStatus
	commander chan AudioCommand
}

func initialModel() model {
	status := make(chan AudioStatus, 1)
	commander := make(chan AudioCommand)

	return model{status, commander}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p":
			m.commander <- playpause
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "fonograpff\n\n"

	status := <-m.status

	// The footer
	s += "Duration: "
	s += status.length.String() + "\n"

	s += "Position: "
	s += status.position.String() + "\n"

	s += "Press p to pause.\n"
	s += "Press q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	model := initialModel()
	p := tea.NewProgram(model)

	err := Play("homies.mp3", model.status, model.commander)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
