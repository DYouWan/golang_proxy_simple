package main

import (
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"proxy/config"
	"proxy/handler"
	"proxy/middleware"
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

		muxHandler, err := NewMuxHandler(cfg.MaxAllowed, cfg.HealthCheck, cfg.HealthCheckInterval, cfg.Routes)
		if err != nil {
			return err
		}

		svr := http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Port),
			Handler: muxHandler,
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

// NewMuxHandler 创建路由处理器 ref: https://github.com/gorilla/mux
func NewMuxHandler(maxAllowed uint,healthCheck bool,healthCheckInterval uint, routing []config.Routing) (*mux.Router,error) {
	muxRouter := mux.NewRouter()
	muxRouter.Use(middleware.PanicsHandling)
	if maxAllowed > 0 {
		muxRouter.Use(middleware.MaxAllowedMiddleware(maxAllowed))
	}

	for _, r := range routing {
		if err := r.ValidationAlgorithm(); err != nil {
			return nil, err
		}
		upstreamPath := r.UpstreamPathParse()
		downstreamPath := r.DownstreamPathParse()
		prefixHandler, err := handler.NewRoutePrefixHandler(r.Algorithm, upstreamPath, downstreamPath, r.DownstreamHosts)
		if err != nil {
			return nil, err
		}

		//每个UpstreamPathTemplate对应多个下游主机，这里判断是否做主机的健康检查
		if healthCheck {
			prefixHandler.HealthCheck(healthCheckInterval)
		}

		//例如上游请求模板配置的是：/apig/config 当请求这个前缀时会匹配对应的RoutePrefixHandler去处理
		muxRouter.PathPrefix(upstreamPath).Handler(prefixHandler).Methods(r.UpstreamHTTPMethod...)

		logging.Infof("Url Path: %s  HTTPMethod:%s 注册成功", upstreamPath, r.UpstreamHTTPMethod)
	}
	return muxRouter, nil
}