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

	git "github.com/go-git/go-git/v5"
)

func run() []error {
	var errs []error

	cwd, err := os.Getwd()
	if err != nil {
		return append(errs, err)
	}
	top, err := filepath.Abs(cwd)
	if err != nil {
		return append(errs, err)
	}
	for {
		if _, err := os.Stat(filepath.Join(top, ".git")); err == nil {
			break
		}

		parent := filepath.Dir(top)
		if parent == top {
			return append(errs, fmt.Errorf("did not find Git repository from filesystem root to %s", cwd))
		}
		top = parent
	}

	repo, err := git.PlainOpen(top)
	if err != nil {
		return append(errs, err)
	}
	tree, err := repo.Worktree()
	if err != nil {
		return append(errs, err)
	}
	status, err := tree.Status()
	if err != nil {
		return append(errs, err)
	}

	wrapErrorForFile := func(file string, err error) error {
		return fmt.Errorf("%s: %v", file, err)
	}

	for fileRepoName, fileStatus := range status {
		// We cannot stat a file that does not exist. We check this upfront.
		if fileStatus.Staging == git.Deleted || fileStatus.Worktree == git.Deleted {
			continue
		}

		fileAbsName := filepath.Join(top, fileRepoName)
		fileBaseName := filepath.Base(fileAbsName)

		// We will need to stat the file anyway, and rather than reading upfront, we will read
		// after the initial deletion check passes.
		fileInfo, err := os.Stat(fileAbsName)
		if err != nil {
			errs = append(errs, wrapErrorForFile(fileRepoName, err))
			continue
		}

		// Logic from upstream "gofmt": https://github.com/golang/go/blob/79bda650410c8617f0ae20dc552c6d5b8f8dcfc8/src/cmd/gofmt/gofmt.go#L76-L80
		if fileInfo.IsDir() || strings.HasPrefix(fileBaseName, ".") || !strings.HasSuffix(fileBaseName, ".go") {
			continue
		}

		content, err := ioutil.ReadFile(fileAbsName)
		if err != nil {
			errs = append(errs, wrapErrorForFile(fileRepoName, err))
			continue
		}
		formatted, err := format.Source(content)
		if err != nil {
			errs = append(errs, wrapErrorForFile(fileRepoName, err))
			continue
		}

		if !bytes.Equal(content, formatted) {
			// We attempt to write to the file with the same permissions it already has.
			if err := ioutil.WriteFile(fileAbsName, formatted, fileInfo.Mode().Perm()); err != nil {
				errs = append(errs, wrapErrorForFile(fileRepoName, err))
				continue
			}
			fmt.Println(fileRepoName)
		}
	}
	return errs
}

func main() {
	flag.Usage = func() {
		fmt.Printf("gofmt-git dev (%s)\nhttps://github.com/nickgerace/gofmt-git\n", runtime.Version())
		flag.PrintDefaults()
	}
	flag.Parse()

	if errs := run(); errs != nil {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
