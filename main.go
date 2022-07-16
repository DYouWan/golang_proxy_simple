package main

import (
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
	"log"
	"net/http"
	"os"
	"proxy/basis/logging"
	"proxy/config"
	"proxy/middleware"
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
	cliApp.Usage = "proxy 1.0 server"
	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "serverConfigFile",
			Value:       "server.yml",
			Destination: &serverConfigFile,
			Usage:       "应用程序配置文件",
		},
		cli.StringFlag{
			Name:        "routeConfigFile",
			Value:       "config.json",
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
		return ServerStart(cfg)
	}

	//Run the CLI app
	if err := cliApp.Run(os.Args); err != nil {
		logging.ERROR.Print(err)
	}
}

func ServerStart(cfg *config.Config) error {
	router := mux.NewRouter()
	router.Use(middleware.PanicsHandling())
	router.Use(middleware.ElapsedTimeHandling())

	for _, route := range cfg.Routes {
		if err := cfg.ValidationAlgorithm(route.Algorithm); err != nil {
			return err
		}
		proxy, err := NewProxy(route)
		if err != nil {
			return err
		}
		if cfg.HealthCheck {
			proxy.HealthCheck(cfg.HealthCheckInterval)
		}
		router.Handle(route.UpstreamPathTemplate, proxy)
	}
	if cfg.MaxAllowed > 0 {
		router.Use(middleware.MaxAllowedMiddleware(cfg.MaxAllowed))
	}
	svr := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: router,
	}

	// listen and serve
	if cfg.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else if cfg.Schema == "https" {
		err := svr.ListenAndServeTLS(cfg.CertCrt, cfg.CertKey)
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	}
	return nil
}