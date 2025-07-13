package ui

import (
	"fmt"
	"strings"

	"claude-squad/session/git"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BranchList is a component for selecting Git branches
type BranchList struct {
	branches           []git.BranchInfo
	selectedIdx        int
	height             int
	width              int
	isRemote           bool
	titleStyle         lipgloss.Style
	selectedStyle      lipgloss.Style
	normalStyle        lipgloss.Style
	currentBranchStyle lipgloss.Style
}

// NewBranchList creates a new branch list component
func NewBranchList(branches []git.BranchInfo, isRemote bool) *BranchList {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color("#dde4f0")).
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#1a1a1a"})

	normalStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.AdaptiveColor{Light: "#444444", Dark: "#cccccc"})

	currentBranchStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#51bd73")).
		Bold(true)

	return &BranchList{
		branches:           branches,
		selectedIdx:        0,
		isRemote:           isRemote,
		titleStyle:         titleStyle,
		selectedStyle:      selectedStyle,
		normalStyle:        normalStyle,
		currentBranchStyle: currentBranchStyle,
	}
}

// SetSize sets the size of the branch list
func (b *BranchList) SetSize(width, height int) {
	b.width = width
	b.height = height
}

// Up moves the selection up
func (b *BranchList) Up() {
	if b.selectedIdx > 0 {
		b.selectedIdx--
	} else {
		b.selectedIdx = len(b.branches) - 1
	}
}

// Down moves the selection down
func (b *BranchList) Down() {
	if b.selectedIdx < len(b.branches)-1 {
		b.selectedIdx++
	} else {
		b.selectedIdx = 0
	}
}

// SelectedBranch returns the currently selected branch
func (b *BranchList) SelectedBranch() *git.BranchInfo {
	if b.selectedIdx >= 0 && b.selectedIdx < len(b.branches) {
		return &b.branches[b.selectedIdx]
	}
	return nil
}

// View renders the branch list
func (b *BranchList) View() string {
	if len(b.branches) == 0 {
		return b.normalStyle.Render("No branches found")
	}

	var sb strings.Builder

	// Title
	title := "Select Local Branch"
	if b.isRemote {
		title = "Select Remote Branch"
	}
	sb.WriteString(b.titleStyle.Render(title))
	sb.WriteString("\n\n")

	// Calculate visible range
	visibleHeight := b.height - 4 // Account for title and padding
	if visibleHeight < 1 {
		visibleHeight = 10 // Default minimum
	}

	startIdx := 0
	endIdx := len(b.branches)

	// Scroll to keep selected item visible
	if b.selectedIdx >= visibleHeight {
		startIdx = b.selectedIdx - visibleHeight + 1
		endIdx = b.selectedIdx + 1
	}
	if endIdx > len(b.branches) {
		endIdx = len(b.branches)
	}
	if endIdx-startIdx > visibleHeight {
		endIdx = startIdx + visibleHeight
	}

	// Render branches
	for i := startIdx; i < endIdx; i++ {
		branch := b.branches[i]
		prefix := "  "
		if branch.IsCurrent && !b.isRemote {
			prefix = "* "
		}

		line := fmt.Sprintf("%s%s", prefix, branch.Name)

		// Apply appropriate style
		var style lipgloss.Style
		if i == b.selectedIdx {
			style = b.selectedStyle
		} else if branch.IsCurrent && !b.isRemote {
			style = b.currentBranchStyle
		} else {
			style = b.normalStyle
		}

		sb.WriteString(style.Render(line))
		if i < endIdx-1 {
			sb.WriteString("\n")
		}
	}

	// Show scroll indicators
	if startIdx > 0 {
		sb.WriteString("\n" + b.normalStyle.Render("  ↑ more branches above"))
	}
	if endIdx < len(b.branches) {
		sb.WriteString("\n" + b.normalStyle.Render("  ↓ more branches below"))
	}

	return sb.String()
}

// Update handles key events for the branch list
func (b *BranchList) Update(msg tea.Msg) (*BranchList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			b.Up()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			b.Down()
		}
	}
	return b, nil
}
