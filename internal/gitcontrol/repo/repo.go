package gitcontrol

import (
	"errors"
	"fmt"
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// GitInfo: 로컬 저장소 관련 정보
type GitInfo struct {
	ID            string
	RepoPath      string // 로컬 git 저장소 경로
	CurrentBranch string
	Repo          *git.Repository // go-git Repository 객체
}

// CommitInfo: 커밋 정보
type CommitInfo struct {
	CommitId   string
	BranchName string
	GitInfoID  string
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

// ----------------------
// 2. 보조 함수
// ----------------------

// (1) 특정 파일에서 ID와 Token 읽기
func LoadAuthFromFile(filePath string) (*Auth, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	// 간단히 첫 줄에 userID, 둘째 줄에 token이 있다고 가정
	if len(lines) < 2 {
		return nil, errors.New("invalid auth file: need at least two lines")
	}

	auth := &Auth{
		UserID: strings.TrimSpace(lines[0]),
		Token:  strings.TrimSpace(lines[1]),
	}
	return auth, nil
}

// (2) 로컬 git 저장소 열기
func OpenGitRepo(repoPath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// (3) 브랜치 설정 (체크아웃)
func CheckoutBranch(gInfo *GitInfo, branchName string) error {
	worktree, err := gInfo.Repo.Worktree()
	if err != nil {
		return err
	}

	// 브랜치가 로컬에 없으면 새로 생성 없이 체크아웃 시도 -> 에러날 수 있음
	// 만약 없으면 자동으로 생성하려면 Create 옵션을 줘야 함
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return err
	}
	gInfo.CurrentBranch = branchName
	return nil
}

// (4) 브랜치 생성
func CreateBranch(gInfo *GitInfo, branchName string) error {
	// 이미 존재하는지 확인
	_, err := gInfo.Repo.Reference(plumbing.NewBranchReferenceName(branchName), false)
	if err == nil {
		return fmt.Errorf("branch %s already exists", branchName)
	}

	// 현재 HEAD 커밋 해시로 새 branch reference 생성
	headRef, err := gInfo.Repo.Head()
	if err != nil {
		return err
	}
	newBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), headRef.Hash())
	if err := gInfo.Repo.Storer.SetReference(newBranchRef); err != nil {
		return err
	}

	return nil
}

// (5) 브랜치 삭제
func DeleteBranch(gInfo *GitInfo, branchName string) error {
	return gInfo.Repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(branchName))
}

// (6) 커밋 수행
func CommitChanges(gInfo *GitInfo, message string) (string, error) {
	worktree, err := gInfo.Repo.Worktree()
	if err != nil {
		return "", err
	}

	// 변경사항 모두 add(간단 처리)
	err = worktree.AddGlob(".")
	if err != nil {
		return "", err
	}

	// 실제 커밋
	commitHash, err := worktree.Commit(message, &git.CommitOptions{
		// Author, committer 등 더 자세히 설정 가능
	})
	if err != nil {
		return "", err
	}

	return commitHash.String(), nil
}

// (7) branch 시작점(최초 커밋)부터 지금까지 변경사항 보기
//   - “시작점”을 어디로 볼 것인가가 관건인데, 여기서는 “branch가 처음 생긴 시점” 또는
//     “branch가 분기된 시점”을 찾는 것이 실제론 복잡합니다.
//   - 예제에선 단순히 “브랜치 상의 전체 커밋 로그”를 순회하면서 변경 파일을 출력하는 식으로 예시를 들겠습니다.
func ShowChangesFromBranchStart(gInfo *GitInfo, branchName string) error {
	refName := plumbing.NewBranchReferenceName(branchName)
	ref, err := gInfo.Repo.Reference(refName, true)
	if err != nil {
		return err
	}

	commitIter, err := gInfo.Repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}
	defer commitIter.Close()

	// 가장 오래된 커밋부터 순회하기 위해, 일단 모든 커밋을 모아서 뒤집는 식으로 예시
	commits := []*object.Commit{}
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return err
	}

	// commits는 최신 순 -> reverse
	for i := 0; i < len(commits)/2; i++ {
		commits[i], commits[len(commits)-1-i] = commits[len(commits)-1-i], commits[i]
	}

	// 이제 순서대로 이력 조회 (커밋 간 diff)
	for i, c := range commits {
		if i == 0 {
			fmt.Printf("----- [%d] Commit %s (Initial in this log) -----\n", i, c.Hash)
			continue
		}

		parent := commits[i-1]
		patch, err := parent.Patch(c)
		if err != nil {
			return err
		}
		fmt.Printf("----- [%d] Commit %s -----\n", i, c.Hash)
		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()
			if to == nil {
				fmt.Printf("  Deleted file: %s\n", from.Path())
				continue
			}
			if from == nil {
				fmt.Printf("  Added file: %s\n", to.Path())
				continue
			}

			changes := filePatch.Chunks()
			for _, chunk := range changes {
				// '+' or '-'
				action := chunk.Type()
				switch action {
				case diff.Operation(merkletrie.Insert):
					fmt.Printf("  [ADDED] in %s\n", to.Path())
				case diff.Operation(merkletrie.Delete):
					fmt.Printf("  [DELETED] in %s\n", from.Path())
				case diff.Operation(merkletrie.Modify):
					fmt.Printf("  [MODIFIED] %s\n", from.Path())
				}
			}
		}
	}

	return nil
}

// (8) branch 내 커밋 확인
func ListCommitsInBranch(gInfo *GitInfo, branchName string) error {
	ref, err := gInfo.Repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		return err
	}
	commitIter, err := gInfo.Repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}
	defer commitIter.Close()

	fmt.Printf("[Commits in branch: %s]\n", branchName)
	commitIter.ForEach(func(c *object.Commit) error {
		fmt.Printf("Commit: %s, Author: %s, Message: %s\n", c.Hash.String()[:7], c.Author.Name, c.Message)
		return nil
	})
	return nil
}

// (9) 특정 커밋ID부터 지금까지 변경사항 보기
func ShowChangesFromCommit(gInfo *GitInfo, commitID string, branchName string) error {
	ref, err := gInfo.Repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		return err
	}
	headCommit, err := gInfo.Repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	startCommit, err := gInfo.Repo.CommitObject(plumbing.NewHash(commitID))
	if err != nil {
		return err
	}

	patch, err := startCommit.Patch(headCommit)
	if err != nil {
		return err
	}

	fmt.Printf("Changes from %s to %s:\n", commitID, headCommit.Hash)
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if to == nil && from != nil {
			fmt.Printf("  Deleted file: %s\n", from.Path())
			continue
		}
		if from == nil && to != nil {
			fmt.Printf("  Added file: %s\n", to.Path())
			continue
		}
		// 수정된 파일 정보
		if from != nil && to != nil && from.Path() == to.Path() {
			fmt.Printf("  Modified file: %s\n", from.Path())
		} else {
			fmt.Printf("  Renamed file: %s -> %s\n", from.Path(), to.Path())
		}
	}

	return nil
}
