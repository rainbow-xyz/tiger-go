// build go1.16

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	cache "saas_service/internal/passport/cache"
	hdv1 "saas_service/internal/passport/delivery/http/v1"
	"saas_service/internal/passport/delivery/http/v1/user"
	repo "saas_service/internal/passport/repository"
	uc "saas_service/internal/passport/usecase"
	"saas_service/internal/pkg/constants"
	"saas_service/internal/pkg/middlewares/ginzap"
	"saas_service/internal/pkg/middlewares/inoauth"
	myrequestid "saas_service/internal/pkg/middlewares/requestid"
	"saas_service/pkg/core"
	"saas_service/pkg/dbm"
	"saas_service/pkg/redism"
	"saas_service/pkg/setting"
	"saas_service/pkg/xlog"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

var tmpKey string

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config := InitConfig()
	xlog.Logger = xlog.InitLogger(config)
	defer xlog.Logger.Sync()
	dbConn := InitDB(config)
	rdbConn := InitRedis(config)

	inoauth.NewXToken(config)
	gin.SetMode(config.ServerCfg.RunMode)

	// todo 通用的收敛到一块
	// Creates a router without any middleware by default
	r := gin.New()
	// Global middleware

	r.Use(requestid.New())
	// 注意先后顺序
	r.Use(myrequestid.SetMyGinRequestID())

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	r.Use(ginzap.GinzapWithConfig(xlog.Logger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        false,
		SkipPaths:  []string{"/system/healthcheck"},
	}))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	r.Use(ginzap.RecoveryWithZap(xlog.Logger, true))

	r.Use(cors.Default())

	// todo 自动依赖注入
	hdv1.NewSystemHandler(r) // passport 系统级别的handler

	// 初始化参数验证器
	core.InitValidator()

	// 业务各层创建
	// user module
	userRepo := repo.NewUserRepo(dbConn)
	userCache := cache.NewUserCache(rdbConn, userRepo)

	// create UserUsecase
	userUcase := uc.NewUserUsecase(userRepo, userCache)
	user.NewUserHandler(r, userUcase)
	srv := NewServer(config, r)

	// todo 整合改进优化 目前先这样  考虑直接封装个app 然后执行app.run()
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func InitConfig() *setting.Config {
	var confFile string = constants.DefaultPassportConfigFilePath

	if path := os.Getenv(constants.PassportConfigFilePathEnvValName); path != "" {
		confFile = path
	}
	// 创建配置文件
	config, err := setting.NewConfig(confFile)
	if err != nil {
		panic(err)
	}
	return config
}

func InitDB(config *setting.Config) *gorm.DB {
	// 连接db
	dbConn, err := dbm.ConnectDB(config.DBCfg)
	if err != nil {
		panic(err)
	}
	return dbConn
}

func InitRedis(config *setting.Config) *redis.Client {
	// 连接redis
	rdbConn, err := redism.ConnectRedis(config.RedisCfg)
	if err != nil {
		panic(err)
	}
	return rdbConn
}

func NewServer(config *setting.Config, h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         ":" + config.ServerCfg.HttpPort,
		Handler:      h,
		ReadTimeout:  config.ServerCfg.ReadTimeout * time.Second,
		WriteTimeout: config.ServerCfg.WriteTimeout * time.Second,
		//MaxHeaderBytes: 1 << 20,
	}
	return srv
}

//func InitLogger(config *setting.Config) *logger
