package session

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"claude-squad/session/git"
	"claude-squad/session/tmux"
)

// BranchInstanceOptions represents options for creating an instance from an existing branch
type BranchInstanceOptions struct {
	Branch  git.BranchInfo
	Path    string
	Program string
	AutoYes bool
}

// NewInstanceFromBranch creates a new instance from an existing branch
func NewInstanceFromBranch(opts BranchInstanceOptions) (*Instance, error) {
	t := time.Now()

	// Convert path to absolute
	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Extract a session title from the branch name
	title := extractTitleFromBranch(opts.Branch.Name)

	// Create the instance
	instance := &Instance{
		Title:     title,
		Status:    Ready,
		Path:      absPath,
		Program:   opts.Program,
		Height:    0,
		Width:     0,
		CreatedAt: t,
		UpdatedAt: t,
		AutoYes:   opts.AutoYes,
		Branch:    opts.Branch.Name, // Store the actual branch name
	}

	return instance, nil
}

// StartFromBranch starts an instance using an existing branch
func (i *Instance) StartFromBranch(branch git.BranchInfo) error {
	if i.Title == "" {
		return fmt.Errorf("instance title cannot be empty")
	}

	// Create tmux session
	var tmuxSession *tmux.TmuxSession
	if i.tmuxSession != nil {
		tmuxSession = i.tmuxSession
	} else {
		tmuxSession = tmux.NewTmuxSession(i.Title, i.Program)
	}
	i.tmuxSession = tmuxSession

	// Create git worktree for the existing branch
	gitWorktree, err := git.NewGitWorktreeForBranch(i.Path, branch.Name)
	if err != nil {
		return fmt.Errorf("failed to create git worktree for branch: %w", err)
	}

	i.gitWorktree = gitWorktree
	i.Branch = branch.Name

	// Setup error handler to cleanup resources on any error
	var setupErr error
	defer func() {
		if setupErr != nil {
			if cleanupErr := i.Kill(); cleanupErr != nil {
				setupErr = fmt.Errorf("%v (cleanup error: %v)", setupErr, cleanupErr)
			}
		} else {
			i.started = true
		}
	}()

	// Checkout the existing branch
	if err := i.gitWorktree.CheckoutExistingBranch(branch.Name, branch.IsRemote); err != nil {
		setupErr = fmt.Errorf("failed to checkout existing branch: %w", err)
		return setupErr
	}

	// Create new session
	if err := i.tmuxSession.Start(i.gitWorktree.GetWorktreePath()); err != nil {
		// Cleanup git worktree if tmux session creation fails
		if cleanupErr := i.gitWorktree.Cleanup(); cleanupErr != nil {
			err = fmt.Errorf("%v (cleanup error: %v)", err, cleanupErr)
		}
		setupErr = fmt.Errorf("failed to start new session: %w", err)
		return setupErr
	}

	i.SetStatus(Running)

	return nil
}

// extractTitleFromBranch converts a branch name to a human-readable session title
func extractTitleFromBranch(branchName string) string {
	// Remove remote prefix if present (e.g., "origin/feature" -> "feature")
	if strings.Contains(branchName, "/") {
		parts := strings.Split(branchName, "/")
		if len(parts) > 1 && (parts[0] == "origin" || parts[0] == "upstream") {
			branchName = strings.Join(parts[1:], "/")
		}
	}

	// Remove common prefixes if they exist
	// This handles cases like "johndoe/fix-bug" -> "fix-bug"
	if idx := strings.LastIndex(branchName, "/"); idx != -1 {
		branchName = branchName[idx+1:]
	}

	// Convert dashes to spaces and capitalize words
	words := strings.Split(branchName, "-")
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
