package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea_ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

const (
	WAV_PATH = "/tmp/goSAM_voice.wav"
)
var (
	History		= make([]string, 5000) 
	Hist_sel	= 0
	Hist_enable	= false

	commands = []string{"/sing", "/quit", "/help"}
)

func main() {
	// Sound Setup
	sampleRate := beep.SampleRate(22050)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func handleInput(s string) {
	if s == "" { return }


	words_s := strings.Split(s, " ")

	switch words_s[0] {
		case "/sing":
			exec.Command("sam", "-sing", "-wav", WAV_PATH, 
						strings.Join(words_s[1:], " ")).Run()
		default:
			exec.Command("sam", "-wav", WAV_PATH, s).Run()
	}
	History = append([]string{s}, History...)


	playSound(WAV_PATH)
}

func playSound(name string) {
	f, err := os.Open(name)
	if err != nil { panic(err) }
	defer f.Close()

	streamer, _, err := wav.Decode(f)
	if err != nil { panic(err) }
	defer streamer.Close()

	ctrl := &beep.Ctrl{
		Streamer: streamer,
	}
	volume := &effects.Volume{
		Streamer:	ctrl,
		Base:		2,
		Volume:		2,
		Silent:		false,
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))

	<-done
}

type errMsg error

type model struct {
	textInput	tea_ti.Model
	err			error
}

func initialModel() model {
	ti := tea_ti.New()
	ti.Placeholder = "Text to speak..."
	ti.Focus()
	ti.Prompt = "goSAM > "
	ti.CharLimit = 512

	return model {
		textInput:	ti,
		err:		nil,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// Main function loop.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() != "" {
				if m.textInput.Value()[0] == '/' {
					comm := strings.Split(m.textInput.Value(), " ")[0]
					switch comm {
					case "/quit":
						return m, tea.Quit
					case "/help":
						fmt.Printf("\n\n/sing\tSAM Sings\n/help\tShows Commands\n/quit\tExit\n")
					}
				}

				fmt.Println("")
				go handleInput(m.textInput.Value())
			}
			m.textInput.Reset()
			Hist_sel = 0
			Hist_enable = false
		case tea.KeyUp:
			if m.textInput.Value() == "" {
				Hist_sel = 0
				m.textInput.SetValue(History[Hist_sel])
				Hist_enable = true
			} else {
				if Hist_enable {
					if Hist_sel < 5000 {
						Hist_sel++
					}
					m.textInput.SetValue(History[Hist_sel])
				}
			}
		case tea.KeyDown:
			if Hist_sel <= 0 {
				m.textInput.Reset()
				Hist_sel = 0
			} else {
				Hist_sel--
				m.textInput.SetValue(History[Hist_sel])
			}
		case tea.KeyTab:
			if m.textInput.Value() != "" && m.textInput.Value()[0] == '/' {
				index := -1
				for i, v := range commands {
					cMem := strings.Contains(v, m.textInput.Value())
					if cMem { index = i }
				}
				m.textInput.SetValue(commands[index])
				m.textInput.SetCursor(len(commands[index]))
			}

		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprint(m.textInput.View()) 
}

