package gitcontrol

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

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

// 브랜치 분기점부터 커밋 내용 비교
func ShowChangesFromBranchPoint(gInfo *GitInfo, branchName, baseBranch string) error {
	ancestor, err := findCommonAncestor(gInfo.Repo, baseBranch, branchName)
	if err != nil {
		return err
	}

	// 2) 현재 브랜치의 HEAD 커밋
	ref, err := gInfo.Repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		return err
	}
	headCommit, err := gInfo.Repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	// 3) 공통 조상 ~ HEAD 간 patch 가져오기
	patch, err := ancestor.Patch(headCommit)
	if err != nil {
		return err
	}

	fmt.Printf("===== Changes from common ancestor %s to HEAD %s (branch: %s) =====\n",
		ancestor.Hash, headCommit.Hash, branchName)

	// 4) 파일별 diff 표시
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()

		// 삭제된 파일
		if from != nil && to == nil {
			fmt.Printf("  Deleted: %s\n", from.Path())
			continue
		}

		// 신규 추가된 파일
		if from == nil && to != nil {
			fmt.Printf("  Added: %s\n", to.Path())
			continue
		}

		// 수정(혹은 rename)된 파일
		if from != nil && to != nil {
			if from.Path() == to.Path() {
				fmt.Printf("  Modified: %s\n", from.Path())
			} else {
				fmt.Printf("  Renamed: %s -> %s\n", from.Path(), to.Path())
			}
		}

		// 추가로 chunk 내용도 보고 싶다면:
		for _, chunk := range filePatch.Chunks() {
			action := chunk.Type()
			switch action {
			case diff.Equal:
				// 변경 없는 부분
			case diff.Add:
				fmt.Printf("    +%s\n", chunk.Content())
			case diff.Delete:
				fmt.Printf("    -%s\n", chunk.Content())
			}
		}
	}

	return nil
}

func findCommonAncestor(repo *git.Repository, branchA, branchB string) (*object.Commit, error) {
	refA, err := repo.Reference(plumbing.NewBranchReferenceName(branchA), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %s: %w", branchA, err)
	}
	refB, err := repo.Reference(plumbing.NewBranchReferenceName(branchB), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %s: %w", branchB, err)
	}

	// branchA에 속한 모든 커밋 해시를 set으로 모으기
	commitsA := make(map[string]bool)

	commitIterA, err := repo.Log(&git.LogOptions{From: refA.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits of %s: %w", branchA, err)
	}
	err = commitIterA.ForEach(func(c *object.Commit) error {
		commitsA[c.Hash.String()] = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	// branchB의 commit을 위로 올라가며, branchA set에 존재하는 것 발견 시 반환
	commitIterB, err := repo.Log(&git.LogOptions{From: refB.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits of %s: %w", branchB, err)
	}

	var commonAncestor *object.Commit
	err = commitIterB.ForEach(func(c *object.Commit) error {
		if commitsA[c.Hash.String()] {
			// 여기서 발견된 커밋이 두 브랜치의 공통 조상
			commonAncestor = c
			// ForEach 내에서 순회를 중단하고 싶으면 에러를 반환하거나 특별한 방법 사용
			return fmt.Errorf("found")
		}
		return nil
	})
	if err != nil && err.Error() != "found" {
		// "found" 가 아닌 다른 에러면 실제 에러로 처리
		return nil, err
	}

	if commonAncestor == nil {
		// 공통 조상 커밋을 못 찾을 수도 있음(히스토리가 완전히 분리된 경우 등)
		return nil, fmt.Errorf("no common ancestor found between %s and %s", branchA, branchB)
	}

	return commonAncestor, nil
}

// branch 내 커밋 확인
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
		fmt.Printf("Commit: %s, Author: %s, Message: %s\n", c.Hash.String(), c.Author.Name, c.Message)
		return nil
	})
	return nil
}
