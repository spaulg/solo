package models

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

type RebuildModel struct {
	soloCtx        *context.CliContext
	projectControl *host.ProjectControl
	progressModel  tea.Model

	width  int
	height int
}

func NewRebuildModel(soloCtx *context.CliContext) (tea.Model, error) {
	projectControl, err := host.ProjectControlFactory(soloCtx)
	if err != nil {
		return nil, err
	}

	progressModel, err := NewProgressModel(soloCtx)
	if err != nil {
		return nil, err
	}

	return &RebuildModel{
		soloCtx:        soloCtx,
		projectControl: projectControl,
		progressModel:  progressModel,
	}, nil
}

func (t *RebuildModel) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		if err := t.projectControl.Rebuild(); err != nil {
			return messages.ErrorMsg(err)
		}

		return ProjectControlActionCompleteMsg{}
	}, t.progressModel.Init())
}

func (t *RebuildModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	//case ExitAfterRefreshMsg:
	//	return t, tea.Quit

	case ProjectControlActionCompleteMsg:
		return t, func() tea.Msg {
			return ExitAfterRefreshMsg{}
		}

	case tea.WindowSizeMsg:
		t.width = m.Width
		t.height = m.Height

		t.progressModel.Update(messages.ModelSizeMsg{
			Width:  t.width,
			Height: t.height,
		})

	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return t, tea.Quit
		}

	default:
		t.progressModel, cmd = t.progressModel.Update(msg)
	}

	return t, cmd
}

func (t *RebuildModel) View() (s string) {
	s += t.progressModel.View()

	return
}
