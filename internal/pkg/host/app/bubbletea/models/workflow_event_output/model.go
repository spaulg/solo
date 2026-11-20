package workflow_event_output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/host/domain"
)

const containerStepMapFile = "container_step_map.meta.json"
const stepMetaFileSuffix = ".meta.json"
const combinedOutputFileSuffix = ".out"

const headerHeight = 1

type Model struct {
	activeModel *layout.ActiveModel

	modelID string
	width   int
	height  int

	auditLogsPath                 string
	containerStepMapRepository    domain.ContainerStepMapRepository
	workflowStepLogMetaRepository domain.WorkflowStepLogMetaRepository
	logReader                     domain.LogReader

	selection *selection
	loading   bool
	err       error

	viewport viewport.Model

	keyMap KeyMap
	styles Styles
}

func DefaultStyles() Styles {
	return Styles{
		StepHeader: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		StepFailed: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")),
		Muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Error:      lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "bottom"),
		),
	}
}

func NewModel(
	activeModel *layout.ActiveModel,
	auditLogsPath string,
	containerStepMapRepository domain.ContainerStepMapRepository,
	workflowStepLogMetaRepository domain.WorkflowStepLogMetaRepository,
	logReader domain.LogReader,
) Model {
	vp := viewport.New()
	vp.SoftWrap = true

	return Model{
		modelID:                       "workflow_event_output",
		activeModel:                   activeModel,
		auditLogsPath:                 auditLogsPath,
		containerStepMapRepository:    containerStepMapRepository,
		workflowStepLogMetaRepository: workflowStepLogMetaRepository,
		logReader:                     logReader,
		viewport:                      vp,
		keyMap:                        DefaultKeyMap(),
		styles:                        DefaultStyles(),
	}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height
		t.viewport.SetWidth(t.width)
		t.viewport.SetHeight(max(0, t.height-headerHeight))

	case messages.ExecutionEventHighlighted, messages.ExecutionEventDeselected:
		t.clear()

	case messages.WorkflowEventSelected:
		sel := selection{
			executionEventName: m.ExecutionEventName,
			workflowName:       m.WorkflowName,
			containerName:      m.ContainerName,
		}

		t.selection = &sel
		t.loading = true
		t.err = nil

		cmd = t.load(sel)

	case workflowEventOutputLoaded:
		// Drop results for a selection that is no longer current
		if t.selection == nil || *t.selection != m.selection {
			break
		}

		t.loading = false
		t.err = m.err
		t.viewport.SetContent(m.content)
		t.viewport.GotoTop()

	case messages.WorkflowEventOutputFocused:
		if t.selection != nil {
			t.activeModel.SetActiveModel(t.modelID)
		}

	case tea.KeyMsg:
		if !t.activeModel.IsActive(t.modelID) {
			break
		}

		switch {
		case key.Matches(m, t.keyMap.Escape):
			cmd = func() tea.Msg {
				return messages.WorkflowEventOutputUnfocused{}
			}

		case key.Matches(m, t.keyMap.Top):
			t.viewport.GotoTop()

		case key.Matches(m, t.keyMap.Bottom):
			t.viewport.GotoBottom()

		default:
			t.viewport, cmd = t.viewport.Update(msg)
		}
	}

	return t, cmd
}

func (t Model) View() tea.View {
	var header, body string

	switch {
	case t.selection == nil:
		return tea.NewView(
			lipgloss.NewStyle().
				Width(t.width).
				MaxWidth(t.width).
				Height(t.height).
				MaxHeight(t.height).
				Align(lipgloss.Center, lipgloss.Center).
				Foreground(lipgloss.Color("241")).
				Render("Select a container workflow run to view its output"),
		)

	case t.loading:
		return tea.NewView(
			lipgloss.NewStyle().
				Width(t.width).
				MaxWidth(t.width).
				Height(t.height).
				MaxHeight(t.height).
				Align(lipgloss.Center, lipgloss.Center).
				Foreground(lipgloss.Color("241")).
				Render("Loading…"),
		)

	case t.err != nil:
		header = t.renderHeader()
		body = t.styles.Error.Render(t.err.Error())

	default:
		header = t.renderHeader()
		body = t.viewport.View()
	}

	return tea.NewView(
		lipgloss.NewStyle().
			Width(t.width).
			MaxWidth(t.width).
			Height(t.height).
			MaxHeight(t.height).
			Render(lipgloss.JoinVertical(lipgloss.Left, header, body)),
	)
}

func (t *Model) clear() {
	t.selection = nil
	t.loading = false
	t.err = nil
	t.viewport.SetContent("")
	t.viewport.GotoTop()
}

func (t Model) renderHeader() string {
	title := t.styles.StepHeader.Render(t.selection.workflowName + " › " + t.selection.containerName)

	var status []string
	if !t.loading && t.err == nil && t.viewport.TotalLineCount() > t.viewport.VisibleLineCount() {
		status = append(status, fmt.Sprintf("%3.f%%", t.viewport.ScrollPercent()*100))
	}

	if t.activeModel.IsActive(t.modelID) {
		status = append(status, "esc to return")
	} else {
		status = append(status, "enter to scroll")
	}

	return lipgloss.NewStyle().MaxWidth(t.width).Render(
		title+"  "+t.styles.Muted.Render(strings.Join(status, " · ")),
	) + "\n"
}

// load reads the combined stderr/stdout output of every step executed for the
// selected container within the selected workflow, in execution order.
func (t Model) load(sel selection) tea.Cmd {
	return func() tea.Msg {
		content, err := t.renderOutput(sel)

		return workflowEventOutputLoaded{
			selection: sel,
			content:   content,
			err:       err,
		}
	}
}

func (t Model) renderOutput(sel selection) (string, error) {
	workflowPath := filepath.Join(t.auditLogsPath, sel.executionEventName, sel.workflowName)

	containerStepMap, err := t.containerStepMapRepository.Load(filepath.Join(workflowPath, containerStepMapFile))
	if err != nil {
		return "", fmt.Errorf("failed to load container step map: %w", err)
	}

	stepIDs := containerStepMap[sel.containerName]
	if len(stepIDs) == 0 {
		return t.styles.Muted.Render("No steps were executed for this container"), nil
	}

	var sb strings.Builder

	for i, stepID := range stepIDs {
		if i > 0 {
			sb.WriteString("\n")
		}

		meta, err := t.workflowStepLogMetaRepository.Load(filepath.Join(workflowPath, stepID+stepMetaFileSuffix))
		if err != nil {
			return "", fmt.Errorf("failed to load step %s meta: %w", stepID, err)
		}

		sb.WriteString(t.renderStepHeader(stepID, meta))
		sb.WriteString("\n")

		output, err := t.readOutput(filepath.Join(workflowPath, stepID+combinedOutputFileSuffix))
		if err != nil {
			return "", fmt.Errorf("failed to read step %s output: %w", stepID, err)
		}

		if output == "" {
			sb.WriteString("\n")
			sb.WriteString(t.styles.Muted.Render("   (no output)"))
			sb.WriteString("\n")

			continue
		}

		if !strings.HasPrefix(output, "\n") {
			sb.WriteString("\n")
		}

		for _, line := range strings.Split(output, "\n") {
			if len(line) > 0 {
				sb.WriteString("   " + line + "\n")
			}
		}
	}

	return sb.String(), nil
}

func (t Model) renderStepHeader(stepID string, meta *domain.WorkflowStepLogMeta) string {
	if meta == nil {
		return t.styles.StepHeader.Render("$ step " + stepID)
	}

	parts := make([]string, 0, len(meta.Arguments)+1)
	parts = append(parts, quoteArgument(meta.Command))

	for _, argument := range meta.Arguments {
		parts = append(parts, quoteArgument(argument))
	}

	commandStyle := t.styles.StepHeader
	if meta.ExitCode != 0 {
		commandStyle = t.styles.StepFailed
	}

	details := fmt.Sprintf("exit: %d", meta.ExitCode)
	if meta.WorkingDirectory != "" {
		details = "path: " + meta.WorkingDirectory + ", " + details
	}

	return commandStyle.Render(" * $ "+strings.Join(parts, " ")) + " " + t.styles.Muted.Render("("+details+")")
}

func (t Model) readOutput(filePath string) (string, error) {
	reader, err := t.logReader.NewReader(filePath)
	if err != nil {
		// Output files are only written when a step produces output
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}

		return "", err
	}

	defer func() {
		_ = reader.Close()
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	output := strings.ReplaceAll(string(data), "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")
	output = strings.TrimSpace(output)

	return output, nil
}

func quoteArgument(argument string) string {
	if argument == "" || strings.ContainsAny(argument, " \t\n\"'\\$`") {
		return strconv.Quote(argument)
	}

	return argument
}
