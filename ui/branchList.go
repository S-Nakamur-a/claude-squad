package ui

import (
	"fmt"
	"strings"

	"claude-squad/session/git"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BranchList is a component for selecting Git branches
type BranchList struct {
	branches           []git.BranchInfo
	filteredBranches   []git.BranchInfo
	selectedIdx        int
	height             int
	width              int
	isRemote           bool
	searchInput        textinput.Model
	titleStyle         lipgloss.Style
	selectedStyle      lipgloss.Style
	normalStyle        lipgloss.Style
	currentBranchStyle lipgloss.Style
	searchStyle        lipgloss.Style
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

	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"})

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Type to filter branches..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	bl := &BranchList{
		branches:           branches,
		filteredBranches:   branches, // Initially show all branches
		selectedIdx:        0,
		isRemote:           isRemote,
		searchInput:        ti,
		titleStyle:         titleStyle,
		selectedStyle:      selectedStyle,
		normalStyle:        normalStyle,
		currentBranchStyle: currentBranchStyle,
		searchStyle:        searchStyle,
	}
	
	return bl
}

// SetSize sets the size of the branch list
func (b *BranchList) SetSize(width, height int) {
	b.width = width
	b.height = height
	if width > 10 {
		b.searchInput.Width = width - 10
	}
}

// Up moves the selection up
func (b *BranchList) Up() {
	if len(b.filteredBranches) == 0 {
		return
	}
	if b.selectedIdx > 0 {
		b.selectedIdx--
	} else {
		b.selectedIdx = len(b.filteredBranches) - 1
	}
}

// Down moves the selection down
func (b *BranchList) Down() {
	if len(b.filteredBranches) == 0 {
		return
	}
	if b.selectedIdx < len(b.filteredBranches)-1 {
		b.selectedIdx++
	} else {
		b.selectedIdx = 0
	}
}

// SelectedBranch returns the currently selected branch
func (b *BranchList) SelectedBranch() *git.BranchInfo {
	if b.selectedIdx >= 0 && b.selectedIdx < len(b.filteredBranches) {
		return &b.filteredBranches[b.selectedIdx]
	}
	return nil
}

// filterBranches filters branches based on search input
func (b *BranchList) filterBranches() {
	searchTerm := strings.ToLower(b.searchInput.Value())
	if searchTerm == "" {
		b.filteredBranches = b.branches
		return
	}

	filtered := make([]git.BranchInfo, 0)
	for _, branch := range b.branches {
		branchNameLower := strings.ToLower(branch.Name)
		// Simple fuzzy search: check if all characters appear in order
		if fuzzyMatch(branchNameLower, searchTerm) {
			filtered = append(filtered, branch)
		}
	}
	
	b.filteredBranches = filtered
	// Reset selection if it's out of bounds
	if b.selectedIdx >= len(b.filteredBranches) {
		b.selectedIdx = 0
	}
}

// fuzzyMatch performs a simple fuzzy match
func fuzzyMatch(text, pattern string) bool {
	patternIdx := 0
	for _, ch := range text {
		if patternIdx >= len(pattern) {
			return true
		}
		if ch == rune(pattern[patternIdx]) {
			patternIdx++
		}
	}
	return patternIdx >= len(pattern)
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
	
	// Search input
	sb.WriteString(b.searchInput.View())
	sb.WriteString("\n")
	sb.WriteString(b.searchStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	sb.WriteString("\n\n")

	// Show message if no branches match the filter
	if len(b.filteredBranches) == 0 {
		sb.WriteString(b.normalStyle.Render("No branches match your search"))
		return sb.String()
	}

	// Calculate visible range
	visibleHeight := b.height - 8 // Account for title, search input, and padding
	if visibleHeight < 1 {
		visibleHeight = 10 // Default minimum
	}

	startIdx := 0
	endIdx := len(b.filteredBranches)

	// Scroll to keep selected item visible
	if b.selectedIdx >= visibleHeight {
		startIdx = b.selectedIdx - visibleHeight + 1
		endIdx = b.selectedIdx + 1
	}
	if endIdx > len(b.filteredBranches) {
		endIdx = len(b.filteredBranches)
	}
	if endIdx-startIdx > visibleHeight {
		endIdx = startIdx + visibleHeight
	}

	// Render branches
	for i := startIdx; i < endIdx; i++ {
		branch := b.filteredBranches[i]
		prefix := "  "
		if branch.IsCurrent && !b.isRemote {
			prefix = "* "
		}

		// Format branch line with last commit time
		var line string
		if branch.LastCommitTime != "" {
			// Calculate padding for alignment
			nameWidth := 40 // Fixed width for branch name column
			paddedName := branch.Name
			if len(paddedName) > nameWidth {
				paddedName = paddedName[:nameWidth-3] + "..."
			} else {
				paddedName = fmt.Sprintf("%-*s", nameWidth, paddedName)
			}
			line = fmt.Sprintf("%s%s  %s", prefix, paddedName, branch.LastCommitTime)
		} else {
			line = fmt.Sprintf("%s%s", prefix, branch.Name)
		}

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
	if endIdx < len(b.filteredBranches) {
		sb.WriteString("\n" + b.normalStyle.Render("  ↓ more branches below"))
	}

	return sb.String()
}

// Update handles key events for the branch list
func (b *BranchList) Update(msg tea.Msg) (*BranchList, tea.Cmd) {
	var cmd tea.Cmd

	// First, update the search input
	prevValue := b.searchInput.Value()
	b.searchInput, cmd = b.searchInput.Update(msg)
	
	// If search value changed, filter branches
	if b.searchInput.Value() != prevValue {
		b.filterBranches()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
			b.Up()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
			b.Down()
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+k"))):
			b.Up()
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+j"))):
			b.Down()
		}
	}
	return b, cmd
}
