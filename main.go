package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "mdns-subdomain"
	app.Usage = "Local mDNS announcer for subdomain"
	app.Flags = flags
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

var flags = []cli.Flag{
	cli.StringFlag{
		EnvVar: "LNAME_IFACE",
		Name:   "iface",
		Usage:  "specify network interface to listen, default listen to all",
	},
	cli.StringFlag{
		EnvVar: "LNAME_HOSTNAME",
		Name:   "hostname",
		Usage:  "specify fixed hostname to broadcast, default the machine hostname",
	},
}

func action(c *cli.Context) error {
	var (
		iface    *net.Interface
		hostname *string
		err      error
	)
	if c.IsSet("iface") {
		iface, err = net.InterfaceByName(c.String("iface"))
		if err != nil {
			return err
		}
	}
	if c.IsSet("hostname") {
		value := c.String("hostname")
		if !strings.HasSuffix(value, ".local") {
			return errors.New("optional hostname must end with .local")
		}
		hostname = &value
	}

	conn, err := listen(iface, hostname)
	if err != nil {
		return err
	}

	gracefulStop := make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		<-gracefulStop
		conn.stop()
	}()
	conn.serve()
	return nil
}
