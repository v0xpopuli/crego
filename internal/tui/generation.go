package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/v0xpopuli/crego/internal/tui/screens"
)

type (
	GenerationStatus int

	GenerationTask struct {
		Key    string
		Label  string
		Status GenerationStatus
		Detail string
	}

	GenerationSummary struct {
		OutputDir string
		FileCount int
		Stack     string
		Elapsed   time.Duration
	}

	GenerationRunner func(func(key string, status GenerationStatus, detail string)) (GenerationSummary, error)

	generationScreen struct {
		styles  Styles
		title   string
		tasks   []GenerationTask
		spin    spinner.Model
		runner  GenerationRunner
		events  chan tea.Msg
		summary GenerationSummary
		err     error
		done    bool
		width   int
		height  int
	}

	generationProgressMsg GenerationTask

	generationDoneMsg struct {
		summary GenerationSummary
		err     error
	}
)

const (
	GenerationPending GenerationStatus = iota
	GenerationRunning
	GenerationDone
	GenerationFailed
)

func NewGenerationApp(title string, tasks []GenerationTask, runner GenerationRunner, opts AppOptions) *App {
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	styles := NewStyles(opts.Out, opts.NoColor)
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	return &App{
		model: NewModel(generationScreen{
			styles: styles,
			title:  title,
			tasks:  tasks,
			spin:   spin,
			runner: runner,
			events: make(chan tea.Msg),
		}, styles),
		in:  opts.In,
		out: opts.Out,
	}
}

func (s generationScreen) Init() tea.Cmd {
	return tea.Batch(s.spin.Tick, s.run(), s.waitProgress())
}

func (s generationScreen) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case generationProgressMsg:
		s.updateTask(GenerationTask(msg))
		return s, s.waitProgress()
	case generationDoneMsg:
		s.done = true
		s.summary = msg.summary
		s.err = msg.err
		if msg.err != nil {
			s.updateTask(GenerationTask{Key: "failed", Label: "Generation failed", Status: GenerationFailed, Detail: msg.err.Error()})
			return s, nil
		}
		return s, nil
	case tea.KeyMsg:
		if s.done && msg.String() == "enter" {
			return s, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spin, cmd = s.spin.Update(msg)
		if s.done {
			return s, nil
		}
		return s, cmd
	}
	return s, nil
}

func (s generationScreen) View() string {
	body := s.taskList()
	help := "generation running"
	if s.done {
		if s.err != nil {
			help = "enter close"
		} else {
			body = s.successView()
			help = "enter close"
		}
	}
	stackLine := ""
	if s.summary.Stack != "" {
		stackLine = s.styles.Footer.Render("Stack: " + s.summary.Stack)
	}
	return RenderShell(s.styles, LayoutProps{
		Title:     s.title,
		Subtitle:  "Generation",
		Body:      body,
		StackLine: stackLine,
		Help:      help,
		Width:     s.width,
		Height:    s.height,
	})
}

func (s generationScreen) UsesShellLayout() bool {
	return true
}

func (s generationScreen) run() tea.Cmd {
	return func() tea.Msg {
		if s.runner == nil {
			return generationDoneMsg{err: fmt.Errorf("generation runner is not configured")}
		}
		summary, err := s.runner(func(key string, status GenerationStatus, detail string) {
			if s.events != nil {
				s.events <- generationProgressMsg{Key: key, Status: status, Detail: detail}
			}
		})
		if s.events != nil {
			close(s.events)
		}
		return generationDoneMsg{summary: summary, err: err}
	}
}

func (s generationScreen) waitProgress() tea.Cmd {
	return func() tea.Msg {
		if s.events == nil {
			return nil
		}
		msg, ok := <-s.events
		if !ok {
			return nil
		}
		return msg
	}
}

func (s *generationScreen) updateTask(task GenerationTask) {
	for index := range s.tasks {
		if s.tasks[index].Key == task.Key {
			if task.Label == "" {
				task.Label = s.tasks[index].Label
			}
			s.tasks[index] = task
			return
		}
	}
	s.tasks = append(s.tasks, task)
}

func (s generationScreen) taskList() string {
	lines := make([]string, 0, len(s.tasks))
	for _, task := range s.tasks {
		marker := "○"
		style := s.styles.Description
		switch task.Status {
		case GenerationRunning:
			marker = s.spin.View()
			style = s.styles.Selected
		case GenerationDone:
			marker = "✓"
			style = s.styles.Success
		case GenerationFailed:
			marker = "!"
			style = s.styles.Error
		}
		line := fmt.Sprintf("%s %s", marker, task.Label)
		if task.Detail != "" {
			line += "  " + s.styles.Description.Render(task.Detail)
		}
		lines = append(lines, style.Render(line))
	}
	return strings.Join(lines, "\n")
}

func (s generationScreen) successView() string {
	lines := []string{
		s.styles.Success.Render("✓ Project generated"),
		fmt.Sprintf("Created: %s", s.summary.OutputDir),
		fmt.Sprintf("Files:   %d", s.summary.FileCount),
		fmt.Sprintf("Stack:   %s", s.summary.Stack),
		fmt.Sprintf("Time:    %.1fs", s.summary.Elapsed.Seconds()),
		"",
		"Next steps",
		fmt.Sprintf("  cd %s", s.summary.OutputDir),
		"  make test",
		"  make run",
	}
	return strings.Join(lines, "\n")
}
