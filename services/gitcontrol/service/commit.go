package gitcontrol

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CommitInfo: 커밋 정보
type CommitInfo struct {
	CommitId   string
	BranchName string
}

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

// (9) 특정 커밋ID부터 지금까지 변경사항 보기
func ShowChangesFromCommit(gInfo *GitInfo, commitInfo *CommitInfo) error {
	ref, err := gInfo.Repo.Reference(plumbing.NewBranchReferenceName(commitInfo.BranchName), true)
	if err != nil {
		return err
	}
	headCommit, err := gInfo.Repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	startCommit, err := gInfo.Repo.CommitObject(plumbing.NewHash(commitInfo.CommitId))
	if err != nil {
		return err
	}

	patch, err := startCommit.Patch(headCommit)
	if err != nil {
		return err
	}

	fmt.Printf("Changes from %s to %s:\n", commitInfo.CommitId, headCommit.Hash)
	for _, filePatch := range patch.FilePatches() {

		from, to := filePatch.Files()
		if to == nil && from != nil {
			fmt.Printf("  [DELETED] in %s\n", from.Path())
			continue
		}
		if from == nil && to != nil {
			fmt.Printf("  [ADDED] in %s\n", to.Path())
			continue
		}
		// 수정된 파일 정보
		if from != nil && to != nil && from.Path() == to.Path() {
			fmt.Printf("  [MODIFIED] in %s\n", from.Path())
		} else {
			fmt.Printf("  [RENAMED] in %s -> %s\n", from.Path(), to.Path())
		}
	}

	return nil
}
