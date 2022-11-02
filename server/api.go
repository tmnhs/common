package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/tmnhs/common"
	"github.com/tmnhs/common/dbclient"
	"github.com/tmnhs/common/etcdclient"
	"github.com/tmnhs/common/logger"
	"github.com/tmnhs/common/notify"
	"github.com/tmnhs/common/redisclient"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	shutdownMaxAge = 15 * time.Second
	shutdownWait   = 1000 * time.Millisecond
)
const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

var (
	ApiOptions struct {
		flags.Options
		Environment       string `short:"e" long:"env" description:"Use ApiServer environment" default:"testing"`
		Version           bool   `short:"v" long:"verbose"  description:"Show ApiServer version"`
		EnablePProfile    bool   `short:"p" long:"enable-pprof"  description:"enable pprof"`
		PProfilePort      int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8188"`
		EnableHealthCheck bool   `short:"a" long:"enable-health-check"  description:"enable health check"`
		HealthCheckURI    string `short:"i" long:"health-check-uri"  description:"health check uri" default:"/health" `
		HealthCheckPort   int    `short:"f" long:"health-check-port"  description:"health check port" default:"8186"`
		ConfigFileName    string `short:"c" long:"config" description:"Use ApiServer config file" default:"main"`
		EnableDevMode     bool   `short:"m" long:"enable-dev-mode"  description:"enable dev mode"`
	}
)

type Option func(c *common.Config)

//注册mysql服务
func WithMysql() Option {
	return func(c *common.Config) {
		mysqlConfig := c.Mysql
		//db
		dsn := mysqlConfig.EmptyDsn()
		createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 ;", mysqlConfig.Dbname)
		if err := dbclient.CreateDatabase(dsn, "mysql", createSql); err != nil {
			logger.GetLogger().Error(fmt.Sprintf("create mysql database failed , error:%s", err.Error()))
		}
		_, err := dbclient.Init(mysqlConfig.Dsn(), mysqlConfig.LogMode, mysqlConfig.MaxIdleConns, mysqlConfig.MaxOpenConns)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("api-server:init mysql failed , error:%s", err.Error()))
		} else {
			logger.GetLogger().Info("api-server:init mysql success")
		}
	}
}

//注册etcd服务
func WithEtcd() Option {
	return func(c *common.Config) {
		etcdConfig := c.Etcd
		//etcd
		_, err := etcdclient.Init(etcdConfig.Endpoints, etcdConfig.DialTimeout, etcdConfig.ReqTimeout)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("api-server:init etcd failed , error:%s", err.Error()))
		} else {
			logger.GetLogger().Info("api-server:init etcd success")
		}
	}
}

//注册通知服务
func WithNotify() Option {
	return func(c *common.Config) {
		//notify
		notify.Init(&notify.Mail{
			Port:     c.Email.Port,
			From:     c.Email.From,
			Host:     c.Email.Host,
			Secret:   c.Email.Secret,
			Nickname: c.Email.Nickname,
		}, &notify.WebHook{
			Url:  c.WebHook.Url,
			Kind: c.WebHook.Kind,
		})

		go notify.Serve()
	}
}

//注册redis服务
func WithRedis() Option {
	return func(c *common.Config) {
		redisConfig := c.Redis
		//reds
		_, err := redisclient.Init(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("api-server:init redis failed , error:%s", err.Error()))
		} else {
			logger.GetLogger().Info("api-server:init redis success")
		}
	}
}

type healthCheckHttpServer struct {
}

func (server *healthCheckHttpServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	io.WriteString(response, "ok\n")
}

var healthCheckServer = &healthCheckHttpServer{}

type ApiServer struct {
	Engine      *gin.Engine
	HttpServer  *http.Server
	Addr        string
	mu          sync.Mutex
	doneChan    chan struct{}
	Routers     []func(*gin.Engine)
	Middlewares []func(*gin.Engine)
	Shutdowns   []func(*ApiServer)
	Services    []func(*ApiServer)
}

//get close Chan
func (srv *ApiServer) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.getDoneChanLocked()
}

func (srv *ApiServer) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *ApiServer) Shutdown(ctx context.Context) {
	//Give priority to business shutdown Hook
	if len(srv.Shutdowns) > 0 {
		for _, shutdown := range srv.Shutdowns {
			shutdown(srv)
		}
	}
	//wait for registry shutdown
	select {
	case <-time.After(shutdownWait):
	}
	// close the HttpServer
	srv.HttpServer.Shutdown(ctx)
}

func (srv *ApiServer) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan /*syscall.SIGUSR1,*/, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()

		for sig := range sigChan {
			if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
				logger.GetLogger().Error(fmt.Sprintf("Graceful shutdown:signal %v to stop api-server ", sig))
				srv.Shutdown(shutdownCtx)
			} else {
				logger.GetLogger().Info(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		logger.Shutdown()
	}()
}

func NewApiServer(opts ...Option) (*ApiServer, error) {
	var parser = flags.NewParser(&ApiOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}

		return nil, err
	}

	if ApiOptions.Version {
		fmt.Printf("%s Version:%s\n", common.ApiModule, common.Version)
		os.Exit(0)
	}

	if ApiOptions.EnablePProfile {
		go func() {
			fmt.Printf("enable pprof http server at:%d\n", ApiOptions.PProfilePort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.PProfilePort), nil))
		}()
	}

	if ApiOptions.EnableHealthCheck {
		go func() {
			fmt.Printf("enable healthcheck http server at:%d\n", ApiOptions.HealthCheckPort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.HealthCheckPort), healthCheckServer))
		}()
	}
	var env = common.Environment(ApiOptions.Environment)
	if env.Invalid() {
		var err error
		env, err = common.NewGlobalEnvironment()
		if err != nil {
			return nil, err
		}
	}

	var configFile = ApiOptions.ConfigFileName
	if configFile == "" {
		configFile = "main"
	}
	defaultConfig, err := common.LoadConfig(env.String(), configFile)
	if err != nil {
		fmt.Printf("api-server:init config error:%s", err.Error())
		return nil, err
	}
	logConfig := defaultConfig.Log
	//log
	logger.Init(logConfig.Level, logConfig.Format, logConfig.Prefix, logConfig.Director, logConfig.ShowLine, logConfig.EncodeLevel, logConfig.StacktraceKey, logConfig.LogInConsole)

	if len(opts) > 0 {
		for _, opt := range opts {
			opt(defaultConfig)
		}
	}
	apiServer := &ApiServer{
		Addr: fmt.Sprintf(":%d", defaultConfig.System.Addr),
	}

	apiServer.setupSignal()
	//set gin mode
	switch env {
	case common.EnvProduction:
		gin.SetMode(gin.ReleaseMode)
	case common.EnvTesting:
		gin.SetMode(gin.DebugMode)
	}
	return apiServer, nil
}

// ListenAndServe Listen And Serve()
func (srv *ApiServer) ListenAndServe() error {
	srv.Engine = gin.New()
	srv.Engine.Use(srv.apiRecoveryMiddleware())
	srv.Engine.Use(srv.cors())

	for _, service := range srv.Services {
		service(srv)
	}

	for _, middleware := range srv.Middlewares {
		middleware(srv.Engine)
	}

	for _, c := range srv.Routers {
		c(srv.Engine)
	}

	srv.HttpServer = &http.Server{
		Handler:        srv.Engine,
		Addr:           srv.Addr,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := srv.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Register Shutdown Handler
func (srv *ApiServer) RegisterShutdown(handlers ...func(*ApiServer)) {
	srv.Shutdowns = append(srv.Shutdowns, handlers...)
}

// Register Service Handler
func (srv *ApiServer) RegisterService(handlers ...func(*ApiServer)) {
	srv.Services = append(srv.Services, handlers...)
}

// Register Middleware Middleware
func (srv *ApiServer) RegisterMiddleware(middlewares ...func(engine *gin.Engine)) {
	srv.Middlewares = append(srv.Middlewares, middlewares...)
}

// RegisterRouters
func (srv *ApiServer) RegisterRouters(routers ...func(engine *gin.Engine)) *ApiServer {
	srv.Routers = append(srv.Routers, routers...)
	return srv
}
