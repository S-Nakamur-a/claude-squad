package overlay

import (
	"claude-squad/session/git"
	"claude-squad/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BranchSelector is an overlay for selecting a Git branch
type BranchSelector struct {
	branchList *ui.BranchList
	width      int
	height     int
	Submitted  bool
	Canceled   bool
	OnSubmit   func(branch git.BranchInfo)
}

// NewBranchSelector creates a new branch selector overlay
func NewBranchSelector(branches []git.BranchInfo, isRemote bool, onSubmit func(branch git.BranchInfo)) *BranchSelector {
	return &BranchSelector{
		branchList: ui.NewBranchList(branches, isRemote),
		OnSubmit:   onSubmit,
	}
}

// Init initializes the component
func (b *BranchSelector) Init() tea.Cmd {
	return nil
}

// SetSize sets the size of the overlay
func (b *BranchSelector) SetSize(width, height int) {
	b.width = width
	b.height = height

	// Calculate inner dimensions for the branch list
	innerWidth := width * 2 / 3
	innerHeight := height * 2 / 3
	if innerWidth > 80 {
		innerWidth = 80
	}
	if innerHeight > 30 {
		innerHeight = 30
	}

	b.branchList.SetSize(innerWidth-4, innerHeight-4) // Account for padding and border
}

// Update handles messages and updates the component
func (b *BranchSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			selected := b.branchList.SelectedBranch()
			if selected != nil && b.OnSubmit != nil {
				b.OnSubmit(*selected)
			}
			b.Submitted = true
			return b, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "ctrl+c"))):
			b.Canceled = true
			return b, nil

		default:
			// Pass other key events to the branch list
			var cmd tea.Cmd
			b.branchList, cmd = b.branchList.Update(msg)
			return b, cmd
		}
	}

	return b, nil
}

// View renders the overlay
func (b *BranchSelector) View() string {
	// Calculate overlay dimensions
	overlayWidth := b.width * 2 / 3
	overlayHeight := b.height * 2 / 3
	if overlayWidth > 80 {
		overlayWidth = 80
	}
	if overlayHeight > 30 {
		overlayHeight = 30
	}

	// Create overlay style
	overlayStyle := lipgloss.NewStyle().
		Width(overlayWidth).
		Height(overlayHeight).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	// Render branch list content
	content := b.branchList.View()

	// Add help text at the bottom
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"}).
		MarginTop(1)

	help := helpStyle.Render("↑/↓/j/k: Navigate • Enter: Select • Esc: Cancel")

	// Combine content and help
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)

	// Apply overlay style
	overlay := overlayStyle.Render(fullContent)

	// Center the overlay
	return lipgloss.Place(
		b.width,
		b.height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
	)
}
