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

const VERSION = "0.0.0-dev"

func run() []error {
	var errors []error

	cwd, err := os.Getwd()
	if err != nil {
		return append(errors, err)
	}
	top, err := filepath.Abs(cwd)
	if err != nil {
		return append(errors, err)
	}
	for {
		if _, err := os.Stat(filepath.Join(top, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(top)
		if parent == top {
			break
		}
		top = parent
	}

	repo, err := git.PlainOpen(top)
	if err != nil {
		return append(errors, err)
	}
	tree, err := repo.Worktree()
	if err != nil {
		return append(errors, err)
	}
	status, err := tree.Status()
	if err != nil {
		return append(errors, err)
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
			errors = append(errors, err)
			continue
		}
		if fileInfo.IsDir() || strings.HasPrefix(fileBaseName, ".") || !strings.HasSuffix(fileBaseName, ".go") {
			continue
		}

		content, err := ioutil.ReadFile(fileAbsName)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		formatted, err := format.Source(content)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// We attempt to write to the file with the same permissions it already has.
		if err := ioutil.WriteFile(fileAbsName, formatted, fileInfo.Mode().Perm()); err != nil {
			errors = append(errors, err)
			continue
		}
		if !bytes.Equal(content, formatted) {
			fmt.Println(fileRepoName)
		}
	}
	return errors
}

func main() {
	flag.Usage = func() {
		fmt.Printf("gofmt-git %s (%s)\nhttps://github.com/nickgerace/gofmt-git\n", VERSION, runtime.Version())
		flag.PrintDefaults()
	}
	flag.Parse()
	if errors := run(); len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}