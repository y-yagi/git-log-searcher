package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/term"
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
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 80
	}

	for _, directory := range config.Directories {
		fmt.Fprintf(outStream, "searching `%v` ...\n", directory)
		searcher := NewSearcher(directory, pattern, width)
		result, err := searcher.Run()
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
		} else {
			fmt.Fprintf(outStream, "%s\n", result)
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
