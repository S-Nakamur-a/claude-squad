package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListLocalBranches(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	err := cmd.Run()
	require.NoError(t, err, "Failed to initialize git repo")
	
	// Create an initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create additional branches
	branches := []string{"feature-1", "feature-2", "bugfix/issue-123"}
	for _, branch := range branches {
		cmd = exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = tempDir
		err = cmd.Run()
		require.NoError(t, err, "Failed to create branch: %s", branch)
	}
	
	// Go back to main branch
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create GitWorktree instance
	worktree := &GitWorktree{
		repoPath: tempDir,
	}
	
	// Test ListLocalBranches
	localBranches, err := worktree.ListLocalBranches()
	require.NoError(t, err)
	
	// Verify we have the expected branches
	assert.Len(t, localBranches, 4) // main + 3 created branches
	
	// Check branch names and that LastCommitTime is populated
	branchNames := make(map[string]bool)
	for _, branch := range localBranches {
		branchNames[branch.Name] = true
		assert.False(t, branch.IsRemote)
		assert.NotEmpty(t, branch.LastCommitTime, "Branch %s should have LastCommitTime", branch.Name)
	}
	
	assert.True(t, branchNames["main"])
	assert.True(t, branchNames["feature-1"])
	assert.True(t, branchNames["feature-2"])
	assert.True(t, branchNames["bugfix/issue-123"])
	
	// Verify current branch is marked correctly
	var currentBranchFound bool
	for _, branch := range localBranches {
		if branch.Name == "main" && branch.IsCurrent {
			currentBranchFound = true
			break
		}
	}
	assert.True(t, currentBranchFound, "Current branch 'main' should be marked as current")
}

func TestListRemoteBranches(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	
	// Create a bare repository to act as remote
	remoteDir := filepath.Join(tempDir, "remote.git")
	cmd := exec.Command("git", "init", "--bare", remoteDir)
	err := cmd.Run()
	require.NoError(t, err, "Failed to create bare repo")
	
	// Create local repository
	localDir := filepath.Join(tempDir, "local")
	err = os.Mkdir(localDir, 0755)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "init")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Add remote
	cmd = exec.Command("git", "remote", "add", "origin", remoteDir)
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create initial commit
	testFile := filepath.Join(localDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Push main branch
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create and push additional branches
	branches := []string{"feature-remote", "release/v1.0"}
	for _, branch := range branches {
		cmd = exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = localDir
		err = cmd.Run()
		require.NoError(t, err)
		
		// Make a commit on each branch (use safe filename)
		safeFilename := strings.ReplaceAll(branch, "/", "_")
		testFile := filepath.Join(localDir, safeFilename+".txt")
		err = os.WriteFile(testFile, []byte("content for "+branch), 0644)
		require.NoError(t, err)
		
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = localDir
		err = cmd.Run()
		require.NoError(t, err)
		
		cmd = exec.Command("git", "commit", "-m", "Commit for "+branch)
		cmd.Dir = localDir
		err = cmd.Run()
		require.NoError(t, err)
		
		cmd = exec.Command("git", "push", "-u", "origin", branch)
		cmd.Dir = localDir
		err = cmd.Run()
		require.NoError(t, err)
	}
	
	// Create GitWorktree instance
	worktree := &GitWorktree{
		repoPath: localDir,
	}
	
	// Test ListRemoteBranches
	remoteBranches, err := worktree.ListRemoteBranches()
	require.NoError(t, err)
	
	// Verify we have the expected branches
	assert.Len(t, remoteBranches, 3) // origin/main, origin/feature-remote, origin/release/v1.0
	
	// Check branch names and properties
	branchNames := make(map[string]bool)
	for _, branch := range remoteBranches {
		branchNames[branch.Name] = true
		assert.True(t, branch.IsRemote)
		assert.False(t, branch.IsCurrent)
	}
	
	assert.True(t, branchNames["origin/main"])
	assert.True(t, branchNames["origin/feature-remote"])
	assert.True(t, branchNames["origin/release/v1.0"])
}

func TestCheckoutExistingBranch_Local(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	err := cmd.Run()
	require.NoError(t, err)
	
	// Create initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create a test branch
	cmd = exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Add a file on the test branch
	branchFile := filepath.Join(tempDir, "branch-file.txt")
	err = os.WriteFile(branchFile, []byte("branch content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "Branch commit")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Go back to main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = tempDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create worktree directory
	worktreeDir := filepath.Join(tempDir, "worktree-test")
	
	// Create GitWorktree instance
	worktree := &GitWorktree{
		repoPath:     tempDir,
		worktreePath: worktreeDir,
		branchName:   "test-branch",
	}
	
	// Test CheckoutExistingBranch for local branch
	err = worktree.CheckoutExistingBranch("test-branch", false)
	require.NoError(t, err)
	
	// Verify the worktree was created
	_, err = os.Stat(worktreeDir)
	require.NoError(t, err, "Worktree directory should exist")
	
	// Verify the branch-specific file exists in the worktree
	worktreeBranchFile := filepath.Join(worktreeDir, "branch-file.txt")
	content, err := os.ReadFile(worktreeBranchFile)
	require.NoError(t, err)
	assert.Equal(t, "branch content", string(content))
	
	// Cleanup worktree
	cmd = exec.Command("git", "worktree", "remove", "-f", worktreeDir)
	cmd.Dir = tempDir
	_ = cmd.Run() // Ignore error in cleanup
}

func TestCheckoutExistingBranch_Remote(t *testing.T) {
	// This test is more complex due to remote branch handling
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	
	// Create a bare repository to act as remote
	remoteDir := filepath.Join(tempDir, "remote.git")
	cmd := exec.Command("git", "init", "--bare", remoteDir)
	err := cmd.Run()
	require.NoError(t, err)
	
	// Create local repository
	localDir := filepath.Join(tempDir, "local")
	err = os.Mkdir(localDir, 0755)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "init")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Add remote
	cmd = exec.Command("git", "remote", "add", "origin", remoteDir)
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create initial commit
	testFile := filepath.Join(localDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Push main branch
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create remote branch
	cmd = exec.Command("git", "checkout", "-b", "remote-feature")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Add file on remote branch
	remoteFile := filepath.Join(localDir, "remote-file.txt")
	err = os.WriteFile(remoteFile, []byte("remote content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "Remote branch commit")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "push", "-u", "origin", "remote-feature")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Delete local branch to simulate fresh checkout
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "branch", "-D", "remote-feature")
	cmd.Dir = localDir
	err = cmd.Run()
	require.NoError(t, err)
	
	// Create worktree directory
	worktreeDir := filepath.Join(localDir, "worktree-remote")
	
	// Create GitWorktree instance
	worktree := &GitWorktree{
		repoPath:     localDir,
		worktreePath: worktreeDir,
		branchName:   "remote-feature", // This will be updated to local name
	}
	
	// Test CheckoutExistingBranch for remote branch
	err = worktree.CheckoutExistingBranch("origin/remote-feature", true)
	require.NoError(t, err)
	
	// Verify the worktree was created
	_, err = os.Stat(worktreeDir)
	require.NoError(t, err, "Worktree directory should exist")
	
	// Verify branch name was updated to local
	assert.Equal(t, "remote-feature", worktree.branchName)
	
	// Verify the remote-specific file exists in the worktree
	worktreeRemoteFile := filepath.Join(worktreeDir, "remote-file.txt")
	content, err := os.ReadFile(worktreeRemoteFile)
	require.NoError(t, err)
	assert.Equal(t, "remote content", string(content))
	
	// Cleanup worktree
	cmd = exec.Command("git", "worktree", "remove", "-f", worktreeDir)
	cmd.Dir = localDir
	_ = cmd.Run() // Ignore error in cleanup
}