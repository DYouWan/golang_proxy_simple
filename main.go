package main

import (
	"github.com/urfave/cli"
	"net/http"
	"os"
	"proxy/config"
	"proxy/util/logging"
	"strconv"
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
	cliApp.Usage = "负载均衡算法：['ip-hash','consistent-hash','p2c','random','round-robin','least-load','bounded']"
	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "serverConfigFile",
			Value:       "config.yml",
			Destination: &serverConfigFile,
			Usage:       "应用程序配置文件",
		},
		cli.StringFlag{
			Name:        "routeConfigFile",
			Value:       "routing.json",
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

		routerHandler, err := NewRouterHandler(cfg)
		if err != nil {
			return err
		}

		svr := http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Port),
			Handler: routerHandler,
		}
		logging.Infof("[%s] proxy 启动成功，正在监听中....", svr.Addr)

		if cfg.Schema == "http" {
			return svr.ListenAndServe()
		} else {
			return svr.ListenAndServeTLS(cfg.CertCrt, cfg.CertKey)
		}
	}

	//运行CLI应用程序
	if err := cliApp.Run(os.Args); err != nil {
		logging.Error(err)
	}
}