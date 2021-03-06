package vcsinfo

import (
	"context"
	"time"
)

// IndexCommit is information about a commit that includes the offset from
// the first commit.
type IndexCommit struct {
	Hash      string
	Index     int
	Timestamp time.Time
}

// ShortCommit stores the hash, author, and subject of a git commit.
type ShortCommit struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Subject string `json:"subject"`
}

// LongCommit gives more detailed information about a commit.
type LongCommit struct {
	*ShortCommit
	Parents   []string        `json:"parent"`
	Body      string          `json:"body"`
	Timestamp time.Time       `json:"timestamp"`
	Branches  map[string]bool `json:"-"`
}

// LongCommitSlice represents a slice of LongCommit objects used for sorting
// commits by timestamp, most recent first.
type LongCommitSlice []*LongCommit

func (s LongCommitSlice) Len() int           { return len(s) }
func (s LongCommitSlice) Less(i, j int) bool { return s[i].Timestamp.After(s[j].Timestamp) }
func (s LongCommitSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// VCS is a generic interface to the information contained in a version
// control system.
type VCS interface {
	// Update updates the local checkout of the repo.
	Update(ctx context.Context, pull, allBranches bool) error

	// From returns commit hashes for the time frame from 'start' to now.
	From(start time.Time) []string

	// Details returns the full commit information for the given hash.
	// If includeBranchInfo is true the Branches field of the returned
	// result will contain all branches that contain the given commit,
	// otherwise Branches will be empty.
	Details(ctx context.Context, hash string, includeBranchInfo bool) (*LongCommit, error)

	// LastNIndex returns the last N commits.
	LastNIndex(N int) []*IndexCommit

	// Range returns all commits from the half open interval ['begin', 'end'), i.e.
	// includes 'begin' and excludes 'end'.
	Range(begin, end time.Time) []*IndexCommit

	// IndexOf returns the index of the commit hash, where 0 is the index of the first hash.
	IndexOf(ctx context.Context, hash string) (int, error)

	// ByIndex returns a LongCommit describing the commit
	// at position N, as ordered in the current branch.
	ByIndex(ctx context.Context, N int) (*LongCommit, error)

	// GetFile returns the content of the given file at the given commit.
	GetFile(ctx context.Context, fileName, commitHash string) (string, error)

	// ResolveCommit will resolve the given commit against a secondary repo if one
	// was defined with the VCS.
	ResolveCommit(ctx context.Context, commitHash string) (string, error)
}
