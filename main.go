package main

import (
	"fmt"
	"os"

	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var (
	name    = "rTorrent XMLRPC CLI"
	version = "1.0.0"
	app     = initApp()
	conn    *rtorrent.RTorrent

	endpoint         string
	view             string
	hash             string
	disableCertCheck bool
)

func initApp() *cli.App {
	nApp := cli.NewApp()

	nApp.Name = name
	nApp.Version = version
	nApp.Authors = []cli.Author{
		{Name: "Michael Robinson", Email: "m@michaelrobinson.io"},
	}

	// Global flags
	nApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "endpoint",
			Usage:       "rTorrent endpoint",
			Value:       "http://myrtorrent/RPC2",
			Destination: &endpoint,
		},
		cli.BoolFlag{
			Name:        "disable-cert-check",
			Usage:       "disable certificate checking on this endpoint, useful for testing",
			Destination: &disableCertCheck,
		},
	}

	nApp.Before = setupConnection

	nApp.Commands = []cli.Command{{
		Name:   "get-ip",
		Usage:  "retrieves the IP for this rTorrent instance",
		Action: getIP,
	}, {
		Name:   "get-name",
		Usage:  "retrieves the name for this rTorrent instance",
		Action: getName,
	}, {
		Name:   "get-totals",
		Usage:  "retrieves the up/down totals for this rTorrent instance",
		Action: getTotals,
	}, {
		Name:   "get-torrents",
		Usage:  "retrieves the torrents from this rTorrent instance",
		Action: getTorrents,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "view",
				Usage:       "view to use, known values: main, started, stopped, hashing, seeding",
				Value:       string(rtorrent.ViewMain),
				Destination: &view,
			},
		},
	}, {
		Name:   "get-files",
		Usage:  "retrieves the files for a specific torrent",
		Action: getFiles,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "hash",
				Usage:       "hash of the torrent",
				Value:       "unknown",
				Destination: &hash,
			},
		},
	},
	}

	return nApp
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func setupConnection(c *cli.Context) error {
	if endpoint == "" {
		return errors.New("endpoint must be specified")
	}
	conn = rtorrent.New(endpoint, disableCertCheck)
	return nil
}

func getIP(c *cli.Context) error {
	ip, err := conn.IP()
	if err != nil {
		return errors.Wrap(err, "failed to get rTorrent IP")
	}
	fmt.Println(ip)
	return nil
}

func getName(c *cli.Context) error {
	name, err := conn.Name()
	if err != nil {
		return errors.Wrap(err, "failed to get rTorrent name")
	}
	fmt.Println(name)
	return nil
}

func getTotals(c *cli.Context) error {
	// Get Down Total
	downTotal, err := conn.DownTotal()
	if err != nil {
		return errors.Wrap(err, "failed to get rTorrent down total")
	}
	fmt.Printf("%d\n", downTotal)

	// Get Up Total
	upTotal, err := conn.UpTotal()
	if err != nil {
		return errors.Wrap(err, "failed to get rTorrent up total")
	}
	fmt.Printf("%d\n", upTotal)
	return nil
}

func getTorrents(c *cli.Context) error {
	torrents, err := conn.GetTorrents(rtorrent.View(view))
	if err != nil {
		return errors.Wrap(err, "failed to get torrents")
	}
	for _, torrent := range torrents {
		fmt.Println(torrent.Pretty())
	}
	return nil
}

func getFiles(c *cli.Context) error {
	files, err := conn.GetFiles(rtorrent.Torrent{Hash: hash})
	if err != nil {
		return errors.Wrap(err, "failed to get files")
	}
	for _, file := range files {
		fmt.Println(file.Pretty())
	}
	return nil
}
