# go-rtorrent
rTorrent XMLRPC Bindings for Go (golang)

## Documentation
[https://godoc.org/github.com/tehjojo/go-rtorrent/rtorrent](https://godoc.org/github.com/tehjojo/go-rtorrent/rtorrent)

## Features
- Get IP, Name, Up/Down totals
- Get torrents within a view
- Get files for torrents
- Set the label on a torrent
- Add a torrent by URL or by metadata

## Installation
To install the package, run `go get github.com/tehjojo/go-rtorrent`

To use it in application, import `"github.com/tehjojo/go-rtorrent/rtorrent"`

To install the command line utility, run `go install "github.com/tehjojo/go-rtorrent"`

## Library Usage

```
conn, _ := rtorrent.New("http://my-rtorrent.com/RPC2", false)
name, _ := conn.Name()
fmt.Printf("My rTorrent's name: %v", name
```

## Command Line Utility
A basic command line utility is included

`go-rtorrent`

```
NAME:
   rTorrent XMLRPC CLI - A new cli application

USAGE:
   go-rtorrent [global options] command [command options] [arguments...]

VERSION:
   0.4.0

AUTHOR(S):
   Michael Robinson <mrobinson@outlook.com>

COMMANDS:
   get-ip    retrieves the IP for this rTorrent instance
   get-name    retrieves the name for this rTorrent instance
   get-totals    retrieves the up/down totals for this rTorrent instance
   get-torrents    retrieves the torrents from this rTorrent instance
   get-files    retrieves the files for a specific torrent
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --endpoint "http://myrtorrent/RPC2"    rTorrent endpoint
   --disable-cert-check            disable certificate checking on this endpoint, useful for testing
   --help, -h                show help
   --version, -v            print the version
```
