package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Searcher struct {
	directory string
	pattern   *regexp.Regexp
	width     int
}

func NewSearcher(directory string, pattern *regexp.Regexp, width int) *Searcher {
	return &Searcher{directory: directory, pattern: pattern, width: width}
}

func (s *Searcher) Run() (string, error) {
	fs := osfs.New(s.directory)
	if _, err := fs.Stat(git.GitDirName); err == nil {
		fs, err = fs.Chroot(git.GitDirName)
		if err != nil {
			return "", err
		}
	}

	storage := filesystem.NewStorageWithOptions(fs, cache.NewObjectLRUDefault(), filesystem.Options{KeepDescriptors: true})
	r, err := git.Open(storage, fs)
	if err != nil {
		return "", err
	}
	defer storage.Close()

	ref, err := r.Head()
	if err != nil {
		return "", err
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return "", err
	}

	result := ""
	err = cIter.ForEach(func(c *object.Commit) error {
		if s.pattern.Match([]byte(c.Message)) {
			result += fmt.Sprintf("%v: %v\n", c.Hash, s.formatMessage(c.Message, s.width-len(c.Hash)))
		}
		return nil
	})

	return result, err
}

func (s *Searcher) formatMessage(msg string, width int) string {
	i := strings.Index(msg, "\n")
	r := []rune(msg)

	if i < width {
		return string(r[0:i])
	}
	return string(r[0:width-3]) + "..."
}
