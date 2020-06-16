# DirectorySync

A simple tool for keeping files and folders in multiple locations synchronized.

## Getting started

### Installing

Download the latest release from the [releases page](https://github.com/j6b72/DirectorySync/releases).

### Prerequisites

- For building, you need to have go [installed and set up correctly](https://golang.org/doc/install).

### Building

- Get the tool `go get github.com/j6b72/DirectorySync`
- Change into the directory `cd $GOPATH/src/github.com/j6b72/DirectorySync`
- Build it `go build -o directorysync`

### Usage

```
Usage: directorysync [options] 

  -h, --help                  Display this help
  -d, --directory <directory> Add a directory to be synchronized with the others
  -c, --config-file <file>    Don't use the configuration.json file and in exchange use the given one
```

A sample configuration file is included in this repository named "configuration.json".

## What the tool can't do (yet?)

- Run constantly in the background and detect changes automatically
- Synchronize deletions

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.