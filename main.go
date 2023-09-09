package main

import (
	"context"
	"ding/initialize/cron"
	"ding/initialize/logger"
	"ding/initialize/mysql"
	"ding/initialize/redis"
	"ding/initialize/viper"
	"ding/routers"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//初始化viper
	err := viper.Init()
	if err != nil {
		zap.L().Error(fmt.Sprintf("init settings failed ,err:%v\n", err))
		return
	}
	zap.L().Debug("viper init success...")
	//初始化Zap
	if err = logger.Init(viper.Conf.LogConfig, viper.Conf.Mode); err != nil {
		zap.L().Error(fmt.Sprintf("init logger failed ,err:%v\n", err))
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("zap init success...")
	//初始化连接飞书
	//global.InitFeishu()
	zap.L().Debug("cron init success...")
	//初始化链接mysql,刚好使用一下gorm，没有用到连表查询，所以比较简单
	if err = mysql.Init(viper.Conf.MySQLConfig); err != nil {
		zap.L().Error(fmt.Sprintf("init mysql failed ,err:%v\n", err))
		return
	}
	//初始化连接redis
	if err = redis.Init(viper.Conf.RedisConfig); err != nil {
		zap.L().Error(fmt.Sprintf("init redis failed ,err:%v\n", err))
		return
	}
	zap.L().Debug("mysql init success...")
	//初始化corn定时器
	if err = cron.InitCorn(); err != nil {
		zap.L().Error(fmt.Sprintf("init cron failed ,err:%v\n", err))
	}

	//将通信201的数据存入数据库
	//if err = gxp.Init(); err != nil {
	//	fmt.Printf("init gxpmysql failed ,err:%v\n", err)
	//	zap.L().Error(fmt.Sprintf("init gxpmysql failed ,err:%v\n", err))
	//	return
	//}
	r := routers.Setup(viper.Conf.Mode)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.Conf.App.Port),
		Handler: r,
	}
	// 初始化kafka
	//if err = initialize.KafkaInit(); err != nil {
	//	zap.L().Error(fmt.Sprintf("kafka init failed ... ,err:%v\n", err))
	//}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("lister: %s\n", err)
			return
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("Server Shutdown", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
