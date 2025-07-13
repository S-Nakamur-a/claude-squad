package session

import (
	"claude-squad/session/git"
)

// BranchInfo re-exports git.BranchInfo for easier access
type BranchInfo = git.BranchInfo

// GetTempWorktreeForBranchListing creates a temporary git worktree for listing branches
func GetTempWorktreeForBranchListing(projectPath string) (*git.GitWorktree, string, error) {
	// We don't need a real session name or branch name for just listing branches
	// Use a placeholder that won't conflict with real sessions
	return git.NewGitWorktree(projectPath, "__temp_branch_listing__")
}
