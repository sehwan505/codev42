package gitcontrol

import (
	"strings"

	git "github.com/go-git/go-git/v5"

	"codev42/configs"
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

// Auth: 인증 정보 (간단히 ID/Token만 관리)
type Auth struct {
	UserID string
	Token  string
}

// (1) 특정 파일에서 ID와 Token 읽기
func LoadAuthFromFile(config *configs.Config) (*Auth, error) {
	auth := &Auth{
		UserID: strings.TrimSpace(config.GitUserID),
		Token:  strings.TrimSpace(config.GitToken),
	}
	return auth, nil
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
