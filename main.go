package main

import (
	"github.com/urfave/cli"
	"os"
	"proxy/basis/logging"
)

var (
	cliApp       *cli.App
	configFile   string
)

func init() {
	cliApp = cli.NewApp()
	cliApp.Name = "proxy-server"
	cliApp.Version = "1.0.0"
	cliApp.Usage = "proxy 1.0 server"
	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "configFile",
			Value:       "config.yml",
			Destination: &configFile,
			Usage:       "configuration file path",
		},
	}
}

func main() {
	cliApp.Action = func(c *cli.Context) error {
		return ServerStart(configFile)
	}

	//Run the CLI app
	if err := cliApp.Run(os.Args); err != nil {
		logging.ERROR.Print(err)
	}
}