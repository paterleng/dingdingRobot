package redis

import (
	"context"
	"ding/global"
	"ding/initialize/viper"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	Perfix     = "ding:"
	ActiveTask = "activeTask:" //活跃任务部分
	Attendance = "attendance:" //考勤状态部分
	User       = "user:"
	UserSign   = User + "sign:"
	LeetCode   = "leetCode:"
)

func Init(redisCfg *viper.RedisConfig) (err error) {
	fmt.Printf("%s,%s,%i", redisCfg.Addr, redisCfg.Password, redisCfg.DB)
	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		zap.L().Error("redis connect ping failed , err :", zap.Error(err))
		return
	} else {
		zap.L().Info("redis connect ping response:", zap.String("pong", pong))
		global.GLOBAL_REDIS = client
		fmt.Println("redis连接成功")
	}
	return
}
