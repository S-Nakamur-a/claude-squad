package ui

import (
	"fmt"
	"testing"

	"claude-squad/session/git"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewBranchList(t *testing.T) {
	branches := []git.BranchInfo{
		{Name: "main", IsCurrent: true, IsRemote: false},
		{Name: "feature-1", IsCurrent: false, IsRemote: false},
	}

	// Test local branch list
	bl := NewBranchList(branches, false)
	assert.NotNil(t, bl)
	assert.Equal(t, branches, bl.branches)
	assert.Equal(t, 0, bl.selectedIdx)
	assert.False(t, bl.isRemote)

	// Test remote branch list
	remoteBranches := []git.BranchInfo{
		{Name: "origin/main", IsCurrent: false, IsRemote: true},
		{Name: "origin/feature", IsCurrent: false, IsRemote: true},
	}
	rbl := NewBranchList(remoteBranches, true)
	assert.NotNil(t, rbl)
	assert.True(t, rbl.isRemote)
}

func TestBranchListNavigation(t *testing.T) {
	branches := []git.BranchInfo{
		{Name: "main", IsCurrent: true, IsRemote: false},
		{Name: "feature-1", IsCurrent: false, IsRemote: false},
		{Name: "feature-2", IsCurrent: false, IsRemote: false},
	}

	bl := NewBranchList(branches, false)

	// Test Down movement
	bl.Down()
	assert.Equal(t, 1, bl.selectedIdx)

	bl.Down()
	assert.Equal(t, 2, bl.selectedIdx)

	// Test wrap around
	bl.Down()
	assert.Equal(t, 0, bl.selectedIdx)

	// Test Up movement
	bl.Up()
	assert.Equal(t, 2, bl.selectedIdx)

	bl.Up()
	assert.Equal(t, 1, bl.selectedIdx)

	bl.Up()
	assert.Equal(t, 0, bl.selectedIdx)
}

func TestBranchListSelectedBranch(t *testing.T) {
	branches := []git.BranchInfo{
		{Name: "main", IsCurrent: true, IsRemote: false},
		{Name: "feature-1", IsCurrent: false, IsRemote: false},
	}

	bl := NewBranchList(branches, false)

	// Test initial selection
	selected := bl.SelectedBranch()
	assert.NotNil(t, selected)
	assert.Equal(t, "main", selected.Name)

	// Test after navigation
	bl.Down()
	selected = bl.SelectedBranch()
	assert.NotNil(t, selected)
	assert.Equal(t, "feature-1", selected.Name)

	// Test with empty list
	emptyBl := NewBranchList([]git.BranchInfo{}, false)
	selected = emptyBl.SelectedBranch()
	assert.Nil(t, selected)
}

func TestBranchListView(t *testing.T) {
	// Test empty branch list
	emptyBl := NewBranchList([]git.BranchInfo{}, false)
	emptyBl.SetSize(50, 20)
	view := emptyBl.View()
	assert.Contains(t, view, "No branches found")

	// Test local branches
	branches := []git.BranchInfo{
		{Name: "main", IsCurrent: true, IsRemote: false},
		{Name: "feature-1", IsCurrent: false, IsRemote: false},
	}

	bl := NewBranchList(branches, false)
	bl.SetSize(50, 20)
	view = bl.View()
	assert.Contains(t, view, "Select Local Branch")
	assert.Contains(t, view, "* main")
	assert.Contains(t, view, "  feature-1")

	// Test remote branches
	remoteBranches := []git.BranchInfo{
		{Name: "origin/main", IsCurrent: false, IsRemote: true},
		{Name: "origin/feature", IsCurrent: false, IsRemote: true},
	}

	rbl := NewBranchList(remoteBranches, true)
	rbl.SetSize(50, 20)
	view = rbl.View()
	assert.Contains(t, view, "Select Remote Branch")
	assert.Contains(t, view, "  origin/main")
	assert.Contains(t, view, "  origin/feature")
}

func TestBranchListUpdate(t *testing.T) {
	branches := []git.BranchInfo{
		{Name: "main", IsCurrent: true, IsRemote: false},
		{Name: "feature-1", IsCurrent: false, IsRemote: false},
		{Name: "feature-2", IsCurrent: false, IsRemote: false},
	}

	bl := NewBranchList(branches, false)

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, cmd := bl.Update(downMsg)
	assert.Nil(t, cmd)
	assert.Equal(t, 1, bl.selectedIdx)

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, cmd = bl.Update(upMsg)
	assert.Nil(t, cmd)
	assert.Equal(t, 0, bl.selectedIdx)

	// Test arrow keys
	downArrow := tea.KeyMsg{Type: tea.KeyDown}
	_, cmd = bl.Update(downArrow)
	assert.Nil(t, cmd)
	assert.Equal(t, 1, bl.selectedIdx)

	upArrow := tea.KeyMsg{Type: tea.KeyUp}
	_, cmd = bl.Update(upArrow)
	assert.Nil(t, cmd)
	assert.Equal(t, 0, bl.selectedIdx)
}

func TestBranchListScrolling(t *testing.T) {
	// Create many branches to test scrolling
	var branches []git.BranchInfo
	for i := 0; i < 20; i++ {
		branches = append(branches, git.BranchInfo{
			Name:      fmt.Sprintf("branch-%d", i),
			IsCurrent: i == 0,
			IsRemote:  false,
		})
	}

	bl := NewBranchList(branches, false)
	bl.SetSize(50, 10) // Small height to force scrolling

	// Navigate to bottom
	for i := 0; i < 15; i++ {
		bl.Down()
	}

	view := bl.View()
	// Should show scroll indicators
	assert.Contains(t, view, "↑ more branches above")
	assert.Contains(t, view, "↓ more branches below")
}

func TestBranchListSizeHandling(t *testing.T) {
	branches := []git.BranchInfo{
		{Name: "main", IsCurrent: true, IsRemote: false},
		{Name: "feature-1", IsCurrent: false, IsRemote: false},
	}

	bl := NewBranchList(branches, false)

	// Test size setting
	bl.SetSize(80, 24)
	assert.Equal(t, 80, bl.width)
	assert.Equal(t, 24, bl.height)

	// Test with very small size
	bl.SetSize(20, 5)
	view := bl.View()
	assert.NotEmpty(t, view) // Should still render something
}