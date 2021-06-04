package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	git "github.com/go-git/go-git/v5"
)

func printFileWrappedError(file string, err error) {
	fmt.Fprintln(os.Stderr, fmt.Errorf("[!] %s: %v", file, err))
}

func processFile(fileRepoName string, fileStatus *git.FileStatus, repositoryRootDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	// We cannot stat a file that does not exist. We check this upfront.
	if fileStatus.Staging == git.Deleted || fileStatus.Worktree == git.Deleted {
		return
	}

	fileAbsName := filepath.Join(repositoryRootDir, fileRepoName)
	fileBaseName := filepath.Base(fileAbsName)

	// We will need to stat the file anyway, and rather than reading upfront, we will read
	// after the initial deletion check passes.
	fileInfo, err := os.Stat(fileAbsName)
	if err != nil {
		printFileWrappedError(fileRepoName, err)
		return
	}

	// Logic from upstream "gofmt": https://github.com/golang/go/blob/79bda650410c8617f0ae20dc552c6d5b8f8dcfc8/src/cmd/gofmt/gofmt.go#L76-L80
	if fileInfo.IsDir() || strings.HasPrefix(fileBaseName, ".") || !strings.HasSuffix(fileBaseName, ".go") {
		return
	}

	content, err := ioutil.ReadFile(fileAbsName)
	if err != nil {
		printFileWrappedError(fileRepoName, err)
		return
	}
	formatted, err := format.Source(content)
	if err != nil {
		printFileWrappedError(fileRepoName, err)
		return
	}

	completed := "[=]"
	if !bytes.Equal(content, formatted) {
		// We attempt to write to the file with the same permissions it already has.
		if err := ioutil.WriteFile(fileAbsName, formatted, fileInfo.Mode().Perm()); err != nil {
			printFileWrappedError(fileRepoName, err)
			return
		}
		completed = "[+]"
	}
	fmt.Printf("%s %s\n", completed, fileRepoName)
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repositoryRootDir, err := filepath.Abs(cwd)
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(repositoryRootDir, ".git")); err == nil {
			break
		}

		parent := filepath.Dir(repositoryRootDir)
		if parent == repositoryRootDir {
			return fmt.Errorf("did not find Git repository from filesystem root to %s", cwd)
		}
		repositoryRootDir = parent
	}

	repo, err := git.PlainOpen(repositoryRootDir)
	if err != nil {
		return err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	status, err := tree.Status()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for fileRepoName, fileStatus := range status {
		wg.Add(1)
		go processFile(fileRepoName, fileStatus, repositoryRootDir, &wg)
	}
	wg.Wait()
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Printf("gofmt-git [dev] (%s)\nhttps://github.com/nickgerace/gofmt-git\n", runtime.Version())
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
