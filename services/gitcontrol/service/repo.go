package gitcontrol

import (
	git "github.com/go-git/go-git/v5"
)

// GitInfo: 로컬 저장소 관련 정보
type GitInfo struct {
	ID            string
	RepoPath      string // 로컬 git 저장소 경로
	CurrentBranch string
	Repo          *git.Repository // go-git Repository 객체
}

// Branch: 브랜치 정보
type Branch struct {
	BranchName string
}

// (2) 로컬 git 저장소 열기
func OpenGitRepo(repoPath string) (*git.Repository, error) {
	if repoPath == "" {
		repoPath = "."
	}
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
