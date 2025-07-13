package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BranchInfo contains information about a Git branch
type BranchInfo struct {
	Name           string
	IsCurrent      bool
	IsRemote       bool
	LastCommitTime string // Relative time of last commit (e.g., "2 hours ago")
}

// ListLocalBranches returns a list of local branches sorted by last commit time
func (g *GitWorktree) ListLocalBranches() ([]BranchInfo, error) {
	// Use git for-each-ref to get branches sorted by committer date
	output, err := g.runGitCommand(g.repoPath, "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)%09%(HEAD)%09%(committerdate:relative)", "refs/heads/")
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	var branches []BranchInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		
		lastCommitTime := ""
		if len(parts) >= 3 {
			lastCommitTime = parts[2]
		}
		
		branches = append(branches, BranchInfo{
			Name:           parts[0],
			IsCurrent:      parts[1] == "*",
			IsRemote:       false,
			LastCommitTime: lastCommitTime,
		})
	}
	return branches, nil
}

// ListRemoteBranches returns a list of remote branches sorted by last commit time
func (g *GitWorktree) ListRemoteBranches() ([]BranchInfo, error) {
	// Use git for-each-ref to get remote branches sorted by committer date
	output, err := g.runGitCommand(g.repoPath, "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)%09%(committerdate:relative)", "refs/remotes/")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	var branches []BranchInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 1 {
			continue
		}
		
		branchName := strings.TrimSpace(parts[0])
		if branchName == "" || strings.Contains(branchName, "HEAD") {
			continue
		}
		
		lastCommitTime := ""
		if len(parts) >= 2 {
			lastCommitTime = parts[1]
		}
		
		branches = append(branches, BranchInfo{
			Name:           branchName,
			IsCurrent:      false,
			IsRemote:       true,
			LastCommitTime: lastCommitTime,
		})
	}
	return branches, nil
}

// CheckoutExistingBranch creates a worktree from an existing branch
func (g *GitWorktree) CheckoutExistingBranch(branchName string, isRemote bool) error {
	// Ensure worktrees directory exists
	worktreesDir := filepath.Join(g.repoPath, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Clean up any existing worktree first
	_, _ = g.runGitCommand(g.repoPath, "worktree", "remove", "-f", g.worktreePath) // Ignore error if worktree doesn't exist

	if isRemote {
		// For remote branches, we need to create a local tracking branch
		// Extract the local branch name from remote (e.g., origin/feature -> feature)
		parts := strings.SplitN(branchName, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid remote branch name: %s", branchName)
		}
		localBranchName := parts[1]

		// Update g.branchName to use the local branch name
		g.branchName = localBranchName

		// Create worktree with tracking branch
		_, err := g.runGitCommand(g.repoPath, "worktree", "add", "-b", localBranchName, g.worktreePath, branchName)
		if err != nil {
			// If branch already exists locally, just create worktree without -b
			_, err = g.runGitCommand(g.repoPath, "worktree", "add", g.worktreePath, localBranchName)
			if err != nil {
				return fmt.Errorf("failed to create worktree from remote branch: %w", err)
			}
		}
	} else {
		// For local branches, just create the worktree
		_, err := g.runGitCommand(g.repoPath, "worktree", "add", g.worktreePath, branchName)
		if err != nil {
			return fmt.Errorf("failed to create worktree from local branch: %w", err)
		}
	}

	return nil
}

