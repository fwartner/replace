# replace
Cli utility for replacing text in files, written in golang and compiled for usage in Docker images

Inspired by https://github.com/piranha/goreplace

## Support me

I invest a lot of resources into creating [best in class open source packages](https://wartner.io/open-source). You can support me by [buying one of my paid products](https://wartner.io/open-source/support-me).

I highly appreciate you sending us a postcard from your hometown, mentioning which of my prpjects you are using. You'll find my address on [my contact page](https://wartner.me/contact). I publish all received postcards on [my virtual postcard wall](https://wartner.io/open-source/postcards).

### Note
This section is inspired by my friends at [Spatie](https://github.com/spatie).

## Features

- Simple search&replace for terms specified as normal shell argument (for escaping only normal shell quotes needed)
- Can use regular expressions for search&replace with and without backrefs (`--regex` and `--regex-backrefs`)
- Supports multiple changesets (search&replace terms)
- Replace the whole line with replacement when line is matching (`--mode=line`)
- ... and add the line at the bottom if there is no match (`--mode=lineinfile`)
- Use [golang template](https://golang.org/pkg/text/template/) with [Sprig template functions]](https://masterminds.github.io/sprig/) (`--mode=template`)
- Can store file as other filename (eg. `replace ./configuration.tmpl:./configuration.conf`)
- Can replace files in directory (`--path`) and offers file pattern matching functions (`--path-pattern` and `--path-regex`)
- Can read also stdin for search&replace or template handling
- Supports Linux, MacOS, Windows and ARM/ARM64 (Rasbperry Pi and others)

## Usage

```
Usage:
  go-replace

Application Options:
      --threads=                                Set thread concurrency for replacing in multiple files at same time (default: 20)
  -m, --mode=[replace|line|lineinfile|template] replacement mode - replace: replace match with term; line: replace line with term; lineinfile: replace line with term or
                                                if not found append to term to file; template: parse content as golang template, search value have to start uppercase
                                                (default: replace)
  -s, --search=                                 search term
  -r, --replace=                                replacement term
      --lineinfile-before=                      add line before this regex
      --lineinfile-after=                       add line after this regex
  -i, --case-insensitive                        ignore case of pattern to match upper and lowercase characters
      --stdin                                   process stdin as input
  -o, --output=                                 write changes to this file (in one file mode)
      --output-strip-ext=                       strip file extension from written files (also available in multi file mode)
      --once=[keep|unique]                      replace search term only one in a file, keep duplicaes (keep, default) or remove them (unique)
      --regex                                   treat pattern as regex
      --regex-backrefs                          enable backreferences in replace term
      --regex-posix                             parse regex term as POSIX regex
      --path=                                   use files in this path
      --path-pattern=                           file pattern (* for wildcard, only basename of file)
      --path-regex=                             file pattern (regex, full path)
      --ignore-empty                            ignore empty file list, otherwise this will result in an error
  -v, --verbose                                 verbose mode
      --dry-run                                 dry run mode
  -V, --version                                 show version and exit
      --dumpversion                             show only version number and exit
  -h, --help                                    show this help message
```

Files must be specified as arguments and will be overwritten after parsing. If you want an alternative location for
saving the file the argument can be specified as `source:destination`, eg.
`go-replace -s foobar -r barfoo daemon.conf.tmpl:daemon.conf`.

If `--path` (with or without `--path-pattern` or `--path-regex`) the files inside path are used as source and will
be overwritten. If `daemon.conf.tmpl` should be written as `daemon.conf` the option `--output-strip-ext=.tmpl` will do
this based on the source file name.

Regular expression's back references can be activated with `--regex-backrefs` and must be specified as `$1, $2 ... $9`.


| Mode       | Description                                                                                                                                                    |
|:-----------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| replace    | Replace search term inside one line with replacement.                                                                                                          |
| line       | Replace line (if matched term is inside) with replacement.                                                                                                     |
| lineinfile | Replace line (if matched term is inside) with replacement. If no match is found in the whole file the line will be appended to the bottom of the file.         |
| template   | Parse content as [golang template](https://golang.org/pkg/text/template/), arguments are available via `{{.Arg.Name}}` or environment vars via `{{.Env.Name}}` |

### Examples

| Command                                                            | Description                                                                                      |
|:-------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------|
| `go-replace -s foobar -r barfoo file1 file2`                       | Replaces `foobar` to `barfoo` in file1 and file2                                                 |
| `go-replace --regex -s 'foo.*' -r barfoo file1 file2`               | Replaces the regex `foo.*` to `barfoo` in file1 and file2                                        |
| `go-replace --regex --ignore-case -s 'foo.*' -r barfoo file1 file2` | Replaces the regex `foo.*` (and ignore case) to `barfoo` in file1 and file2                      |
| `go-replace --mode=line -s 'foobar' -r barfoo file1 file2`          | Replaces all lines with content `foobar` to `barfoo` (whole line) in file1 and file2             |
| `go-replace -s 'foobar' -r barfoo --path=./ --path-pattern='*.txt'` | Replaces all lines with content `foobar` to `barfoo` (whole line) in *.txt files in current path |

### Example with golang templates

Withing the template there are [Template functions available from Sprig](https://masterminds.github.io/sprig/).

Configuration file `daemon.conf.tmpl`:
```
<VirtualHost ...>
    ServerName {{env "SERVERNAME"}}
    DocumentRoot {{env "DOCUMENTROOT"}}
<VirtualHost>

```

Process file with:

```bash
export SERVERNAME=www.foobar.example
export DOCUMENTROOT=/var/www/foobar.example/
go-replace --mode=template daemon.conf.tmpl:daemon.conf
```

Reuslt file `daemon.conf`:
```
<VirtualHost ...>
    ServerName www.foobar.example
    DocumentRoot /var/www/foobar.example/
<VirtualHost>
```

## Installation

```bash
REPLACE_VERSION=1.0.0 \
&& wget -O /usr/local/bin/replace https://github.com/fwartner/replace/releases/download/REPLACE_VERSION/gr-64-linux \
&& chmod +x /usr/local/bin/replace
```

## Docker images

| Image                          | Description                                     |
|:-------------------------------|:------------------------------------------------|
| `wartnerio/replace:latest`  | Latest release, binary only                     |
| `wartnerio/replace:master`  | Current development version in branch `master`  |
| `wartnerio/replace:develop` | Current development version in branch `develop` |
  
 If you like what I am doing please consider [sponsor my work](https://github.com/sponsors/fwartner)!