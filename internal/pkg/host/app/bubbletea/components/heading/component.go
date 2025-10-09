package heading

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
)

type Component struct {
	heading string
	style   lipgloss.Style

	width  int
	height int
}

func DefaultStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("212"))
}

func WithStyle(style lipgloss.Style) Option {
	return func(component *Component) {
		component.style = style
	}
}

func NewComponent(heading string, opts ...Option) Component {
	component := Component{
		heading: heading,
		style:   DefaultStyle(),
	}

	for _, opt := range opts {
		opt(&component)
	}

	return component
}

func (t Component) Init() tea.Cmd {
	return nil
}

func (t Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height
	}

	return t, nil
}

func (t Component) View() tea.View {
	return tea.NewView(t.style.Width(t.width).Height(t.height).Render(t.heading))
}
