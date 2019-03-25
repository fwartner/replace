package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func writeContentToFile(fileitem fileitem, content bytes.Buffer) (string, bool) {
	if opts.DryRun {
		return content.String(), true
	} else {
		var err error
		err = ioutil.WriteFile(fileitem.Output, content.Bytes(), 0644)
		if err != nil {
			panic(err)
		}

		return fmt.Sprintf("%s found and replaced match\n", fileitem.Path), true
	}
}

func searchFilesInPath(path string, callback func(os.FileInfo, string)) {
	var pathRegex *regexp.Regexp

	if opts.PathRegex != "" {
		pathRegex = regexp.MustCompile(opts.PathRegex)
	}

	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		filename := f.Name()

		if f.IsDir() {
			if contains(pathFilterDirectories, f.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if opts.PathPattern != "" {
			matched, _ := filepath.Match(opts.PathPattern, filename)
			if !matched {
				return nil
			}
		}

		if pathRegex != nil {
			if !pathRegex.MatchString(path) {
				return nil
			}
		}

		callback(f, path)
		return nil
	})
}
