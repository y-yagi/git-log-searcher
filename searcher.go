package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

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
	outStream io.Writer
	errStream io.Writer
}

func NewSearcher(directory string, pattern *regexp.Regexp, width int, outStream, errStream io.Writer) *Searcher {
	return &Searcher{directory: directory, pattern: pattern, width: width, outStream: outStream, errStream: errStream}
}

func (s *Searcher) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	fs := osfs.New(s.directory)
	if _, err := fs.Stat(git.GitDirName); err == nil {
		fs, err = fs.Chroot(git.GitDirName)
		if err != nil {
			fmt.Fprintf(s.errStream, "error occured in `%s`: %s", s.directory, err)
			return
		}
	}

	storage := filesystem.NewStorageWithOptions(fs, cache.NewObjectLRUDefault(), filesystem.Options{KeepDescriptors: true})
	r, err := git.Open(storage, fs)
	if err != nil {
		fmt.Fprintf(s.errStream, "error occured in `%s`: %s", s.directory, err)
		return
	}
	defer storage.Close()

	ref, err := r.Head()
	if err != nil {
		fmt.Fprintf(s.errStream, "error occured in `%s`: %s", s.directory, err)
		return
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		fmt.Fprintf(s.errStream, "error occured in `%s`: %s", s.directory, err)
		return
	}

	result := ""
	err = cIter.ForEach(func(c *object.Commit) error {
		if s.pattern.Match([]byte(c.Message)) {
			result += fmt.Sprintf("%v: %v\n", c.Hash, s.formatMessage(c.Message, s.width-len(c.Hash)))
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(s.errStream, "error occured in `%s`: %s", s.directory, err)
		return
	}

	fmt.Fprintf(s.outStream, "searched `%v`\n%s\n", s.directory, result)
}

func (s *Searcher) formatMessage(msg string, width int) string {
	i := strings.Index(msg, "\n")
	r := []rune(msg)

	if i < width {
		return string(r[0:i])
	}
	return string(r[0:width-3]) + "..."
}
