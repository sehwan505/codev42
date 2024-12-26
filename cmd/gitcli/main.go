package main

import (
	"fmt"
	"os"

	gitcontrol "github.com/sehwan505/codev42/internal/gitcontrol/repo"

	"github.com/spf13/cobra"
)

var gInfo = &gitcontrol.GitInfo{}
var authInfo = &gitcontrol.Auth{}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gitcli",
		Short: "Simple git cli with go-git",
	}

	// 1) 특정 파일에서 ID/Token 불러오기
	loadCmd := &cobra.Command{
		Use:   "load-auth [filePath]",
		Short: "Load userID and token from file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			a, err := gitcontrol.LoadAuthFromFile(filePath)
			if err != nil {
				return err
			}
			authInfo = a
			fmt.Printf("Loaded Auth: userID=%s, token=%s\n", authInfo.UserID, authInfo.Token)
			return nil
		},
	}

	// 2) git repository 연동 (열기)
	repoCmd := &cobra.Command{
		Use:   "open-repo [path]",
		Short: "Open local git repo from path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := args[0]
			repo, err := gitcontrol.OpenGitRepo(repoPath)
			if err != nil {
				return err
			}
			gInfo.RepoPath = repoPath
			gInfo.Repo = repo

			// 현재 브랜치 파악
			headRef, err := repo.Head()
			if err != nil {
				return err
			}
			branchName := headRef.Name().Short()
			gInfo.CurrentBranch = branchName
			fmt.Printf("Opened repo at %s (current branch: %s)\n", repoPath, branchName)
			return nil
		},
	}

	// 3) branch 설정하기(체크아웃)
	checkoutCmd := &cobra.Command{
		Use:   "checkout [branchName]",
		Short: "Checkout to branchName",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			err := gitcontrol.CheckoutBranch(gInfo, branchName)
			if err != nil {
				return err
			}
			fmt.Printf("Now on branch: %s\n", gInfo.CurrentBranch)
			return nil
		},
	}

	// 4) branch 생성하기
	createBranchCmd := &cobra.Command{
		Use:   "create-branch [branchName]",
		Short: "Create a new branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			err := gitcontrol.CreateBranch(gInfo, branchName)
			if err != nil {
				return err
			}
			fmt.Printf("Branch '%s' created.\n", branchName)
			return nil
		},
	}

	// 5) branch 지우기
	deleteBranchCmd := &cobra.Command{
		Use:   "delete-branch [branchName]",
		Short: "Delete branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			err := gitcontrol.DeleteBranch(gInfo, branchName)
			if err != nil {
				return err
			}
			fmt.Printf("Branch '%s' deleted.\n", branchName)
			return nil
		},
	}

	// 6) commit 하기
	commitCmd := &cobra.Command{
		Use:   "commit [message]",
		Short: "Commit all changes with given message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := args[0]
			hash, err := gitcontrol.CommitChanges(gInfo, message)
			if err != nil {
				return err
			}
			fmt.Printf("Committed with hash: %s\n", hash)
			return nil
		},
	}

	// 7) branch 시작점부터 지금까지 변경사항 확인
	showBranchChangesCmd := &cobra.Command{
		Use:   "show-branch-changes [branchName]",
		Short: "Show changes from branch start to HEAD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			err := gitcontrol.ShowChangesFromBranchStart(gInfo, branchName)
			return err
		},
	}

	// 8) branch 내 커밋 확인
	listCommitsCmd := &cobra.Command{
		Use:   "list-commits [branchName]",
		Short: "List commits in a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			return gitcontrol.ListCommitsInBranch(gInfo, branchName)
		},
	}

	// 9) 특정 커밋ID부터 지금까지 변경사항 확인
	showChangesFromCommitCmd := &cobra.Command{
		Use:   "show-changes [commitID] [branchName]",
		Short: "Show changes from commitID to HEAD of branch",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			commitID := args[0]
			branchName := args[1]
			return gitcontrol.ShowChangesFromCommit(gInfo, commitID, branchName)
		},
	}

	// 명령어 등록
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(checkoutCmd)
	rootCmd.AddCommand(createBranchCmd)
	rootCmd.AddCommand(deleteBranchCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(showBranchChangesCmd)
	rootCmd.AddCommand(listCommitsCmd)
	rootCmd.AddCommand(showChangesFromCommitCmd)

	// 실행
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
