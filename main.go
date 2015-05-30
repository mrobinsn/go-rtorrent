package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/michaeltrobinson/go-rtorrent/rtorrent"
)

var (
	name    = "rTorrent XMLRPC CLI"
	version = "0.0.1"
	app     = initApp()
	conn    *rtorrent.RTorrent
)

func initApp() *cli.App {
	app := cli.NewApp()

	app.Name = name
	app.Version = version
	app.Authors = []cli.Author{
		{Name: "Michael Robinson", Email: "mrobinson@outlook.com"},
	}

	// Global flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "endpoint",
			Usage: "rTorrent endpoint",
			Value: "http://myrtorrent/RPC2",
		},
		cli.BoolFlag{
			Name:  "disable-cert-check",
			Usage: "disable certificate checking on this endpoint, useful for testing",
		},
	}

	app.Before = setupConnection

	app.Commands = []cli.Command{{
		Name:   "get-ip",
		Usage:  "retrieves the IP for this rTorrent instance",
		Action: getIP,
		Before: setupConnection,
	}, {
		Name:   "get-name",
		Usage:  "retrieves the name for this rTorrent instance",
		Action: getName,
		Before: setupConnection,
	}, {
		Name:   "get-totals",
		Usage:  "retrieves the up/down totals for this rTorrent instance",
		Action: getTotals,
		Before: setupConnection,
	},
	}

	return app
}

func main() {
	app.Run(os.Args)
}

func setupConnection(c *cli.Context) error {
	rTorrentConn, err := rtorrent.New(c.GlobalString("endpoint"), c.GlobalBool("disable-cert-check"))
	if err != nil {
		fmt.Printf("[ERR] Error creating rTorrent connection: %v\n", err)
	}

	conn = rTorrentConn

	return err
}

func getIP(c *cli.Context) {
	ip, err := conn.IP()
	if err != nil {
		fmt.Printf("[ERR] Error getting rTorrent IP: %v\n", err)
	} else {
		fmt.Printf("[INFO] rTorrent IP: %v\n", *ip)
	}
}

func getName(c *cli.Context) {
	name, err := conn.Name()
	if err != nil {
		fmt.Printf("[ERR] Error getting rTorrent name: %v\n", err)
	} else {
		fmt.Printf("[INFO] rTorrent name: %v\n", *name)
	}
}

func getTotals(c *cli.Context) {
	// Get Down Total
	downTotal, err := conn.DownTotal()
	if err != nil {
		fmt.Printf("[ERR] Error getting rTorrent down total: %v\n", err)
	} else {
		fmt.Printf("[INFO] rTorrent down total: %v bytes\n", downTotal)
	}

	// Get Up Total
	upTotal, err := conn.UpTotal()
	if err != nil {
		fmt.Printf("[ERR] Error getting rTorrent up total: %v\n", err)
	} else {
		fmt.Printf("[INFO] rTorrent up total: %v bytes\n", upTotal)
	}
}
