package main

import (
	"fmt"
	"os"

	"github.com/sehwan505/codev42/configs"
	"github.com/sehwan505/codev42/internal/gitcontrol"
	"github.com/spf13/cobra"
)

var gInfo = &gitcontrol.GitInfo{}
var authInfo = &gitcontrol.Auth{}
var config = &configs.Config{}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gitcli",
		Short: "git cli for codev42",
	}
	config, err := configs.GetConfig()
	if err != nil {
		return
	}
	auth, err := gitcontrol.LoadAuthFromFile(config)
	if err != nil {
		return
	}
	authInfo = auth
	fmt.Printf("Loaded Auth: userID=%s, token=%s\n", authInfo.UserID, authInfo.Token)

	repo, err := gitcontrol.OpenGitRepo(config.GitRepo)
	if err != nil {
		return
	}
	gInfo.RepoPath = config.GitRepo
	gInfo.Repo = repo

	// 현재 브랜치 파악
	headRef, err := repo.Head()
	if err != nil {
		return
	}
	branchName := headRef.Name().Short()
	gInfo.CurrentBranch = branchName
	fmt.Printf("Opened repo at %s (current branch: %s)\n", gInfo.Repo, branchName)

	// branch 설정하기(체크아웃)
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

	// branch 생성하기
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

	// branch 지우기
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

	// commit 하기
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

	// branch 시작점부터 지금까지 변경사항 확인
	showBranchChangesCmd := &cobra.Command{
		Use:   "show-branch-changes [branchName] [baseBranch]",
		Short: "Show changes from branch start to HEAD",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			baseBranch := args[1]
			err := gitcontrol.ShowChangesFromBranchPoint(gInfo, branchName, baseBranch)
			return err
		},
	}

	// branch 내 커밋 확인
	listCommitsCmd := &cobra.Command{
		Use:   "list-commits [branchName]",
		Short: "List commits in a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			return gitcontrol.ListCommitsInBranch(gInfo, branchName)
		},
	}

	// 특정 커밋ID부터 지금까지 변경사항 확인
	showChangesFromCommitCmd := &cobra.Command{
		Use:   "show-changes [commitID] [branchName]",
		Short: "Show changes from commitID to HEAD of branch",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			commitID := args[0]
			branchName := args[1]
			commitInfo := &gitcontrol.CommitInfo{commitID, branchName}
			return gitcontrol.ShowChangesFromCommit(gInfo, commitInfo)
		},
	}

	// 명령어 등록
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
