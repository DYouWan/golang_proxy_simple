package main

import (
	"github.com/urfave/cli"
	"os"
	"proxy/basis/logging"
	"proxy/config"
)

var (
	cliApp           *cli.App
	routeConfigFile  string
	serverConfigFile string
)

func init() {
	cliApp = cli.NewApp()
	cliApp.Name = "proxy-server"
	cliApp.Version = "1.0.0"
	cliApp.Usage = "proxy 1.0 server"
	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "serverConfigFile",
			Value:       "config/config.yml",
			Destination: &serverConfigFile,
			Usage:       "应用程序配置文件",
		},
		cli.StringFlag{
			Name:        "routeConfigFile",
			Value:       "config/routes.json",
			Destination: &routeConfigFile,
			Usage:       "路由配置文件",
		},
	}
}

func main() {
	cliApp.Action = func(c *cli.Context) error {
		files := []string{serverConfigFile, routeConfigFile}
		cfg, err := config.Read(true, files...)
		if err != nil {
			return err
		}
		server := NewServer(cfg)
		return server.Start()
	}

	//Run the CLI app
	if err := cliApp.Run(os.Args); err != nil {
		logging.ERROR.Print(err)
	}
}