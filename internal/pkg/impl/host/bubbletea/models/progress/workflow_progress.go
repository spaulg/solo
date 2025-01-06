package progress

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
)

type WorkflowProgressModel struct {
	spinner      spinner.Model
	workflowName workflowcommon.WorkflowName
	status       progresscommon.ComposeProgressStatus
	output       string
}

func NewWorkflowProgressModel() tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))

	return &WorkflowProgressModel{
		spinner: s,
	}
}

func (t *WorkflowProgressModel) Init() tea.Cmd {
	return t.spinner.Tick
}

func (t *WorkflowProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case spinner.TickMsg:
		if t.status == progresscommon.Complete {
			return t, nil
		}

		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd

	case messages.WorkflowStartedMsg:
		t.workflowName = m.WorkflowName
		t.status = progresscommon.InProgress

	case messages.WorkflowStepStartedMsg:
	case messages.WorkflowStepOutputMsg:
		// todo: messages dont guarantee that they'll deliver in order
		//		 so need a better solution when considering delivery of
		//		 output in multiple messages
		//		 how about re-reading the output from the log files

	case messages.WorkflowStepCompleteMsg:

	case messages.WorkflowCompleteMsg:
		t.status = progresscommon.Complete
	}

	return t, nil
}

func (t *WorkflowProgressModel) View() (s string) {
	if t.status == progresscommon.InProgress {
		s += t.spinner.View()
	} else {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
		s += style.Render("\u2714 ")
	}

	s += "Running " + t.workflowName.String()

	//if t.status == progresscommon.InProgress {
	//s += t.output
	//}

	return
}
