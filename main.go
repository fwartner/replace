package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"github.com/remeh/sizedwaitgroup"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

const (
	Author  = "wartner.io"
	Version = "1.0.0"
)

type changeset struct {
	SearchPlain string
	Search      *regexp.Regexp
	Replace     string
	MatchFound  bool
}

type changeresult struct {
	File   fileitem
	Output string
	Status bool
	Error  error
}

type fileitem struct {
	Path   string
	Output string
}

var opts struct {
	ThreadCount        int    `           long:"threads"                       description:"Set thread concurrency for replacing in multiple files at same time" default:"20"`
	Mode               string `short:"m"  long:"mode"                          description:"replacement mode - replace: replace match with term; line: replace line with term; lineinfile: replace line with term or if not found append to term to file; template: parse content as golang template, search value have to start uppercase" default:"replace" choice:"replace" choice:"line" choice:"lineinfile" choice:"template"`
	ModeIsReplaceMatch bool
	ModeIsReplaceLine  bool
	ModeIsLineInFile   bool
	ModeIsTemplate     bool
	Search             []string `short:"s"  long:"search"                        description:"search term"`
	Replace            []string `short:"r"  long:"replace"                       description:"replacement term"`
	LineinfileBefore   string   `           long:"lineinfile-before"             description:"add line before this regex"`
	LineinfileAfter    string   `           long:"lineinfile-after"              description:"add line after this regex"`
	CaseInsensitive    bool     `short:"i"  long:"case-insensitive"              description:"ignore case of pattern to match upper and lowercase characters"`
	Stdin              bool     `           long:"stdin"                         description:"process stdin as input"`
	Output             string   `short:"o"  long:"output"                        description:"write changes to this file (in one file mode)"`
	OutputStripFileExt string   `           long:"output-strip-ext"              description:"strip file extension from written files (also available in multi file mode)"`
	Once               string   `           long:"once"                          description:"replace search term only one in a file, keep duplicaes (keep, default) or remove them (unique)" optional:"true" optional-value:"keep" choice:"keep" choice:"unique"`
	Regex              bool     `           long:"regex"                         description:"treat pattern as regex"`
	RegexBackref       bool     `           long:"regex-backrefs"                description:"enable backreferences in replace term"`
	RegexPosix         bool     `           long:"regex-posix"                   description:"parse regex term as POSIX regex"`
	Path               string   `           long:"path"                          description:"use files in this path"`
	PathPattern        string   `           long:"path-pattern"                  description:"file pattern (* for wildcard, only basename of file)"`
	PathRegex          string   `           long:"path-regex"                    description:"file pattern (regex, full path)"`
	IgnoreEmpty        bool     `           long:"ignore-empty"                  description:"ignore empty file list, otherwise this will result in an error"`
	Verbose            bool     `short:"v"  long:"verbose"                       description:"verbose mode"`
	DryRun             bool     `           long:"dry-run"                       description:"dry run mode"`
	ShowVersion        bool     `short:"V"  long:"version"                       description:"show version and exit"`
	ShowOnlyVersion    bool     `           long:"dumpversion"                   description:"show only version number and exit"`
	ShowHelp           bool     `short:"h"  long:"help"                          description:"show this help message"`
}

var pathFilterDirectories = []string{"autom4te.cache", "blib", "_build", ".bzr", ".cdv", "cover_db", "CVS", "_darcs", "~.dep", "~.dot", ".git", ".hg", "~.nib", ".pc", "~.plst", "RCS", "SCCS", "_sgbak", ".svn", "_obj", ".idea"}

func applyChangesetsToFile(fileitem fileitem, changesets []changeset) (string, bool, error) {
	var (
		err    error  = nil
		output string = ""
		status bool   = true
	)

	file, err := os.Open(fileitem.Path)
	if err != nil {
		return output, false, err
	}

	writeBufferToFile := false
	var buffer bytes.Buffer

	r := bufio.NewReader(file)
	line, e := Readln(r)
	for e == nil {
		newLine, lineChanged, skipLine := applyChangesetsToLine(line, changesets)

		if lineChanged || skipLine {
			writeBufferToFile = true
		}

		if !skipLine {
			buffer.WriteString(newLine + "\n")
		}

		line, e = Readln(r)
	}
	file.Close()

	if opts.ModeIsLineInFile {
		lifBuffer, lifStatus := handleLineInFile(changesets, buffer)
		if lifStatus {
			buffer.Reset()
			buffer.WriteString(lifBuffer.String())
			writeBufferToFile = lifStatus
		}
	}

	if opts.Output != "" || opts.OutputStripFileExt != "" {
		writeBufferToFile = true
	}

	if writeBufferToFile {
		output, status = writeContentToFile(fileitem, buffer)
	} else {
		output = fmt.Sprintf("%s no match", fileitem.Path)
	}

	return output, status, err
}

func applyTemplateToFile(fileitem fileitem, changesets []changeset) (string, bool, error) {
	var (
		err    error  = nil
		output string = ""
		status bool   = true
	)

	buffer, err := ioutil.ReadFile(fileitem.Path)
	if err != nil {
		return output, false, err
	}

	content := parseContentAsTemplate(string(buffer), changesets)

	output, status = writeContentToFile(fileitem, content)

	return output, status, err
}

func applyChangesetsToLine(line string, changesets []changeset) (string, bool, bool) {
	changed := false
	skipLine := false

	for i, changeset := range changesets {
		if opts.Once != "" && changeset.MatchFound {
			if opts.Once == "unique" && searchMatch(line, changeset) {
				skipLine = true
				changed = true
				break
			}
		} else {
			if searchMatch(line, changeset) {
				if opts.ModeIsReplaceLine || opts.ModeIsLineInFile {
					if opts.RegexBackref {
						line = string(changeset.Search.Find([]byte(line)))
						line = changeset.Search.ReplaceAllString(line, changeset.Replace)
					} else {
						line = changeset.Replace
					}
				} else {
					line = replaceText(line, changeset)
				}

				changesets[i].MatchFound = true
				changed = true
			}
		}
	}

	return line, changed, skipLine
}

func buildSearchTerm(term string) *regexp.Regexp {
	var ret *regexp.Regexp
	var regex string

	if opts.Regex {
		regex = term
	} else {
		regex = regexp.QuoteMeta(term)
	}

	if opts.CaseInsensitive {
		regex = "(?i:" + regex + ")"
	}

	if opts.Verbose {
		logMessage(fmt.Sprintf("Using regular expression: %s", regex))
	}

	if opts.RegexPosix {
		ret = regexp.MustCompilePOSIX(regex)
	} else {
		ret = regexp.MustCompile(regex)
	}

	return ret
}

func handleSpecialCliOptions(args []string) {
	if opts.ShowOnlyVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if opts.ShowVersion {
		fmt.Println(fmt.Sprintf("replace version %s", Version))
		fmt.Println(fmt.Sprintf("Copyright (C) 2019 %s", Author))
		os.Exit(0)
	}

	if opts.ShowHelp {
		argparser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	switch mode := opts.Mode; mode {
	case "replace":
		opts.ModeIsReplaceMatch = true
		opts.ModeIsReplaceLine = false
		opts.ModeIsLineInFile = false
		opts.ModeIsTemplate = false
	case "line":
		opts.ModeIsReplaceMatch = false
		opts.ModeIsReplaceLine = true
		opts.ModeIsLineInFile = false
		opts.ModeIsTemplate = false
	case "lineinfile":
		opts.ModeIsReplaceMatch = false
		opts.ModeIsReplaceLine = false
		opts.ModeIsLineInFile = true
		opts.ModeIsTemplate = false
	case "template":
		opts.ModeIsReplaceMatch = false
		opts.ModeIsReplaceLine = false
		opts.ModeIsLineInFile = false
		opts.ModeIsTemplate = true
	}

	if opts.Output != "" && len(args) > 1 {
		logFatalErrorAndExit(errors.New("Only one file is allowed when using --output"), 1)
	}

	if opts.LineinfileBefore != "" || opts.LineinfileAfter != "" {
		if !opts.ModeIsLineInFile {
			logFatalErrorAndExit(errors.New("--lineinfile-after and --lineinfile-before only valid in --mode=lineinfile"), 1)
		}

		if opts.LineinfileBefore != "" && opts.LineinfileAfter != "" {
			logFatalErrorAndExit(errors.New("Only --lineinfile-after or --lineinfile-before is allowed in --mode=lineinfile"), 1)
		}
	}
}

func actionProcessStdinReplace(changesets []changeset) int {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		newLine, _, skipLine := applyChangesetsToLine(line, changesets)

		if !skipLine {
			fmt.Println(newLine)
		}
	}

	return 0
}

func actionProcessStdinTemplate(changesets []changeset) int {
	var buffer bytes.Buffer

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + "\n")
	}

	content := parseContentAsTemplate(buffer.String(), changesets)
	fmt.Print(content.String())

	return 0
}

func actionProcessFiles(changesets []changeset, fileitems []fileitem) int {
	if len(fileitems) == 0 {
		if opts.IgnoreEmpty {
			logMessage("No files found, requsted to ignore this")
			os.Exit(0)
		} else {
			logFatalErrorAndExit(errors.New("No files specified"), 1)
		}
	}

	swg := sizedwaitgroup.New(8)
	results := make(chan changeresult, len(fileitems))

	for _, file := range fileitems {
		swg.Add()
		go func(file fileitem, changesets []changeset) {
			var (
				err    error  = nil
				output string = ""
				status bool   = true
			)

			if opts.ModeIsTemplate {
				output, status, err = applyTemplateToFile(file, changesets)
			} else {
				output, status, err = applyChangesetsToFile(file, changesets)
			}

			results <- changeresult{file, output, status, err}
			swg.Done()
		}(file, changesets)
	}

	swg.Wait()
	close(results)

	errorCount := 0
	for result := range results {
		if result.Error != nil {
			logError(result.Error)
			errorCount++
		} else if opts.Verbose {
			title := fmt.Sprintf("%s:", result.File.Path)

			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, title)
			fmt.Fprintln(os.Stderr, strings.Repeat("-", len(title)))
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, result.Output)
			fmt.Fprintln(os.Stderr, "")
		}
	}

	if errorCount >= 1 {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("[ERROR] %s failed with %d error(s)", argparser.Command.Name, errorCount))
		return 1
	}

	return 0
}

func buildChangesets() []changeset {
	var changesets []changeset

	if !opts.ModeIsTemplate {
		if len(opts.Search) == 0 || len(opts.Replace) == 0 {
			logFatalErrorAndExit(errors.New("Missing either --search or --replace for this mode"), 1)
		}
	}

	if len(opts.Search) != len(opts.Replace) {
		logFatalErrorAndExit(errors.New("Unequal numbers of search or replace options"), 1)
	}

	for i := range opts.Search {
		search := opts.Search[i]
		replace := opts.Replace[i]

		changeset := changeset{search, buildSearchTerm(search), replace, false}
		changesets = append(changesets, changeset)
	}

	return changesets
}

func buildFileitems(args []string) []fileitem {
	var (
		fileitems []fileitem
		file      fileitem
	)

	for _, filepath := range args {
		file = fileitem{filepath, filepath}

		if opts.Output != "" {
			file.Output = opts.Output
		} else if opts.OutputStripFileExt != "" {
			file.Output = strings.TrimSuffix(file.Output, opts.OutputStripFileExt)
		} else if strings.Contains(filepath, ":") {
			split := strings.SplitN(filepath, ":", 2)

			file.Path = split[0]
			file.Output = split[1]
		}

		fileitems = append(fileitems, file)
	}

	if opts.Path != "" {
		searchFilesInPath(opts.Path, func(f os.FileInfo, filepath string) {
			file := fileitem{filepath, filepath}

			if opts.OutputStripFileExt != "" {
				file.Output = strings.TrimSuffix(file.Output, opts.OutputStripFileExt)
			}

			fileitems = append(fileitems, file)
		})
	}

	return fileitems
}

var argparser *flags.Parser

func main() {
	argparser = flags.NewParser(&opts, flags.PassDoubleDash)
	args, err := argparser.Parse()

	handleSpecialCliOptions(args)

	if err != nil {
		logFatalErrorAndExit(err, 1)
	}

	changesets := buildChangesets()
	fileitems := buildFileitems(args)

	exitMode := 0
	if opts.Stdin {
		if opts.ModeIsTemplate {
			exitMode = actionProcessStdinTemplate(changesets)
		} else {
			exitMode = actionProcessStdinReplace(changesets)
		}
	} else {
		exitMode = actionProcessFiles(changesets, fileitems)
	}

	os.Exit(exitMode)
}
