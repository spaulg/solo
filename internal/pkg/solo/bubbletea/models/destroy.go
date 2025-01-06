package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/solo/context"
)

type DestroyModel struct {
	soloCtx               *context.CliContext
	projectControl        *solo.ProjectControl
	projectActionComplete bool
	progressModel         tea.Model
}

func NewDestroyModel(soloCtx *context.CliContext) (*DestroyModel, error) {
	projectControl, err := solo.ProjectControlFactory(soloCtx)
	if err != nil {
		return nil, err
	}

	progressModel, err := NewProgressModel(soloCtx)
	if err != nil {
		return nil, err
	}

	return &DestroyModel{
		soloCtx:        soloCtx,
		projectControl: projectControl,
		progressModel:  *progressModel,
	}, nil
}

func (t DestroyModel) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		if err := t.projectControl.Destroy(); err != nil {
			return messages.ErrorMsg(err)
		}

		if err := t.projectControl.Clean(false); err != nil {
			return messages.ErrorMsg(err)
		}

		return ProjectControlActionCompleteMsg{}
	}, t.progressModel.Init())
}

func (t DestroyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case ExitAfterRefreshMsg:
		return t, tea.Quit

	case ProjectControlActionCompleteMsg:
		t.projectActionComplete = true
		return t, func() tea.Msg {
			return ExitAfterRefreshMsg{}
		}

	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return t, tea.Quit
		}

	default:
		t.progressModel, cmd = t.progressModel.Update(msg)
	}

	return t, cmd
}

func (t DestroyModel) View() (s string) {
	s += t.progressModel.View()

	if t.projectActionComplete {
		s += "\n\nProject successfully destroyed\n"
	}

	return
}
