package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
	"github.com/concertim/flight-user-suite/flight/pkg"
)

// sessionState is used to track which model is focused
type sessionState uint

const (
	guideView sessionState = iota
	spinnerView
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	modelStyle = lipgloss.NewStyle().
			Width(40).
			Height(10).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(40).
				Height(10).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(pkg.LightBlue)
	spinnerStyle = lipgloss.NewStyle().Foreground(pkg.AlcesBlue)
	helpStyle    = lipgloss.NewStyle().Foreground(pkg.Grey)
)

type mainModel struct {
	state   sessionState
	guide   example
	spinner spinner.Model
	index   int
}

type example struct {
	viewport viewport.Model
}

func newExample(isDark bool) (*example, error) {
	const (
		width  = 38
		height = 10
	)

	vp := viewport.New()
	vp.SetWidth(width)
	vp.SetHeight(height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(pkg.AlcesBlue)

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//
	const glamourGutter = 3
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	markdownThemeFilename := filepath.Join(markdownThemeDir, "flight-light.json")
	if isDark {
		markdownThemeFilename = filepath.Join(markdownThemeDir, "flight-dark.json")
	}
	markdownThemeBytes, err := os.ReadFile(markdownThemeFilename)
	if err != nil {
		return nil, err
	}
	var style ansi.StyleConfig
	if err := json.Unmarshal(markdownThemeBytes, &style); err != nil {
		return nil, fmt.Errorf("parsing markdown theme: %w", err)
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}

	markdown, err := markdownContent()
	if err != nil {
		return nil, err
	}
	str, err := renderer.Render(string(markdown))
	if err != nil {
		return nil, err
	}

	vp.SetContent(str)

	return &example{
		viewport: vp,
	}, nil
}

func markdownContent() ([]byte, error) {
	filenames, err := collectMarkdownFiles(howtoDir)
	if err != nil {
		return nil, fmt.Errorf("collecting guide files: %w", err)
	}
	howtoName := filenames[1]
	fullPath := filepath.Join(howtoDir, howtoName)
	if !strings.HasSuffix(fullPath, ".md") {
		fullPath = fullPath + ".md"
	}
	markdown, err := os.ReadFile(fullPath)
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); ok {
			if pathError.Err.Error() == "no such file or directory" {
				return nil, UnknownHowto{Howto: howtoName}
			}
		}
		return nil, fmt.Errorf("reading howto: %w", err)
	}
	return markdown, nil
}

func newModel() (mainModel, error) {
	m := mainModel{state: guideView}
	guide, err := newExample(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	m.guide = *guide
	m.spinner = spinner.New()
	return m, err
}

func (m mainModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.state == guideView {
				m.state = spinnerView
			} else {
				m.state = guideView
			}
		}
		switch m.state {
		// update whichever model is focused
		case spinnerView:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		default:
			m.guide.viewport, cmd = m.guide.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() tea.View {
	var s strings.Builder
	model := m.currentFocusedModel()
	if m.state == guideView {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, focusedModelStyle.Render(fmt.Sprintf("%4s", m.guide.viewport.View())), modelStyle.Render(m.spinner.View())))
	} else {
		s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(fmt.Sprintf("%4s", m.guide.viewport.View())), focusedModelStyle.Render(m.spinner.View())))
	}
	s.WriteString(helpStyle.Render(fmt.Sprintf("\ntab: focus next • n: new %s • q: exit\n", model)))
	return tea.NewView(s.String())
}

func (m mainModel) currentFocusedModel() string {
	if m.state == guideView {
		return "guide"
	}
	return "spinner"
}

func (m *mainModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *mainModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func tui() {
	var p *tea.Program
	if m, err := newModel(); err == nil {
		p = tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}
