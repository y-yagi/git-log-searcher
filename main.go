package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/pelletier/go-toml/v2"
)

const (
	app = "git-log-searcher"
)

var (
	flags *flag.FlagSet
)

type Config struct {
	Directories []string `toml:directories`
}

func setFlags() {
	flags = flag.NewFlagSet(app, flag.ExitOnError)
}

func main() {
	setFlags()
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, outStream, errStream io.Writer) (exitCode int) {
	flags.Parse(os.Args[1:])

	if flags.NArg() != 1 {
		fmt.Fprintln(errStream, "please specify pattern")
		return 1
	}

	config, err := parseDataFile("git-log-searcher.toml")
	if err != nil {
		fmt.Fprintf(errStream, "config file parse error: %v\n", err)
		return 1
	}

	pattern := regexp.MustCompile(flags.Args()[0])

	for _, directory := range config.Directories {
		fmt.Fprintf(outStream, "searching `%v` ...\n", directory)
		fs := osfs.New(directory)
		if _, err := fs.Stat(git.GitDirName); err == nil {
			fs, err = fs.Chroot(git.GitDirName)
			if err != nil {
				fmt.Fprintf(errStream, "error: %v\n", err)
				return 1
			}
		}
		s := filesystem.NewStorageWithOptions(fs, cache.NewObjectLRUDefault(), filesystem.Options{KeepDescriptors: true})
		r, err := git.Open(s, fs)
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			return 1
		}
		defer s.Close()

		ref, err := r.Head()
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			return 1
		}

		cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			return 1
		}

		err = cIter.ForEach(func(c *object.Commit) error {
			if pattern.Match([]byte(c.Message)) {
				fmt.Fprintf(outStream, "%v: %v\n-----------------------\n", c.Hash, c.Message)
			}
			return nil
		})

		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			return 1
		}
	}

	return 0
}

func parseDataFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = toml.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
