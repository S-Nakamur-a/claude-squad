package session

import (
	"os"
	"testing"

	"claude-squad/log"
	"claude-squad/session/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize the logger before any tests run
	log.Initialize(false)
	defer log.Close()

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestNewInstanceFromBranch(t *testing.T) {
	tests := []struct {
		name       string
		opts       BranchInstanceOptions
		wantTitle  string
		wantBranch string
		wantErr    bool
	}{
		{
			name: "local branch",
			opts: BranchInstanceOptions{
				Branch: git.BranchInfo{
					Name:      "feature-test",
					IsCurrent: false,
					IsRemote:  false,
				},
				Path:    ".",
				Program: "test-program",
				AutoYes: false,
			},
			wantTitle:  "Feature Test",
			wantBranch: "feature-test",
			wantErr:    false,
		},
		{
			name: "remote branch",
			opts: BranchInstanceOptions{
				Branch: git.BranchInfo{
					Name:      "origin/feature-remote",
					IsCurrent: false,
					IsRemote:  true,
				},
				Path:    ".",
				Program: "test-program",
				AutoYes: true,
			},
			wantTitle:  "Feature Remote",
			wantBranch: "origin/feature-remote",
			wantErr:    false,
		},
		{
			name: "branch with prefix",
			opts: BranchInstanceOptions{
				Branch: git.BranchInfo{
					Name:      "johndoe/fix-bug-123",
					IsCurrent: false,
					IsRemote:  false,
				},
				Path:    ".",
				Program: "test-program",
			},
			wantTitle:  "Fix Bug 123",
			wantBranch: "johndoe/fix-bug-123",
			wantErr:    false,
		},
		{
			name: "branch with multiple slashes",
			opts: BranchInstanceOptions{
				Branch: git.BranchInfo{
					Name:      "release/v1.0/hotfix",
					IsCurrent: false,
					IsRemote:  false,
				},
				Path:    ".",
				Program: "test-program",
			},
			wantTitle:  "Hotfix",
			wantBranch: "release/v1.0/hotfix",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := NewInstanceFromBranch(tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, instance)
			assert.Equal(t, tt.wantTitle, instance.Title)
			assert.Equal(t, tt.wantBranch, instance.Branch)
			assert.Equal(t, Ready, instance.Status)
			assert.Equal(t, tt.opts.Program, instance.Program)
			assert.Equal(t, tt.opts.AutoYes, instance.AutoYes)
		})
	}
}

func TestExtractTitleFromBranch(t *testing.T) {
	tests := []struct {
		branchName string
		expected   string
	}{
		// Simple cases
		{"feature-test", "Feature Test"},
		{"fix-bug", "Fix Bug"},
		{"main", "Main"},
		
		// Remote branches
		{"origin/feature-test", "Feature Test"},
		{"upstream/fix-bug", "Fix Bug"},
		
		// Branches with user prefixes
		{"johndoe/feature-test", "Feature Test"},
		{"alice/fix-bug-123", "Fix Bug 123"},
		
		// Complex cases
		{"origin/johndoe/feature-test", "Feature Test"},
		{"release/v1.0/hotfix", "Hotfix"},
		{"feature/JIRA-123/implement-login", "Implement Login"},
		
		// Edge cases
		{"feature", "Feature"},
		{"feature-", "Feature "},
		{"-feature", " Feature"},
		{"feature--test", "Feature  Test"},
		
		// No dashes
		{"develop", "Develop"},
		{"master", "Master"},
		
		// Mixed case preservation
		{"feature-OAuth-integration", "Feature OAuth Integration"},
	}

	for _, tt := range tests {
		t.Run(tt.branchName, func(t *testing.T) {
			result := extractTitleFromBranch(tt.branchName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: Testing StartFromBranch requires more complex setup with git repositories
// and tmux sessions, which would be better suited for integration tests.
// The test below provides a basic structure that could be expanded with proper mocking.

func TestStartFromBranch_Validation(t *testing.T) {
	// Create instance with empty title
	instance := &Instance{
		Title: "",
	}

	branch := git.BranchInfo{
		Name:     "test-branch",
		IsRemote: false,
	}

	// Should fail with empty title
	err := instance.StartFromBranch(branch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance title cannot be empty")
}

// Additional integration test structure (would require proper setup)
func TestStartFromBranch_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This test would require:
	// 1. Setting up a real git repository
	// 2. Creating test branches
	// 3. Mocking or using real tmux sessions
	// 4. Verifying worktree creation
	// 5. Cleaning up resources

	// Example structure:
	/*
	// Setup test git repo
	tempDir := t.TempDir()
	// ... initialize git repo, create branches ...

	// Create instance
	instance := &Instance{
		Title:   "Test Session",
		Path:    tempDir,
		Program: "echo 'test'",
	}

	// Create branch info
	branch := git.BranchInfo{
		Name:     "test-branch",
		IsRemote: false,
	}

	// Mock tmux session if needed
	// ...

	// Test starting from branch
	err := instance.StartFromBranch(branch)
	require.NoError(t, err)

	// Verify worktree was created
	// Verify tmux session was started
	// Verify status is Running

	// Cleanup
	err = instance.Kill()
	require.NoError(t, err)
	*/
}