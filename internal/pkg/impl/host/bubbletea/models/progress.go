package models

import (
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/models/progress"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

type ProjectControlActionCompleteMsg struct{}
type ExitAfterRefreshMsg struct{}

type ProgressModel struct {
	soloCtx            *context.CliContext
	infrastructureTree tea.Model
	width              int
	height             int
}

func NewProgressModel(soloCtx *context.CliContext) (tea.Model, error) {
	return &ProgressModel{
		soloCtx:            soloCtx,
		infrastructureTree: progress.NewInfrastructureTreeModel(),
	}, nil
}

func (t *ProgressModel) Init() tea.Cmd {
	return nil
}

func (t *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case messages.ModelSizeMsg:
		t.width = m.Width
		t.height = m.Height
	}

	t.infrastructureTree.Update(msg)

	return t, nil
}

func (t *ProgressModel) View() string {
	leftPanelWidth := int(math.Ceil((float64(t.width) / 100) * 20))
	rightPanelWidth := t.width - leftPanelWidth

	s1 := lipgloss.NewStyle().
		Width(leftPanelWidth).
		Height(t.height).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("228")).
		BorderBackground(lipgloss.Color("63")).
		BorderRight(true).
		Render(t.infrastructureTree.View())

	s2 := lipgloss.NewStyle().
		Width(rightPanelWidth).
		Height(t.height).
		Padding(1, 2).
		Render("Output")

	return lipgloss.JoinHorizontal(lipgloss.Top, s1, s2)
}
