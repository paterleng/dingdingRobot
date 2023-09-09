package cron

import (
	"ding/global"
	"ding/model/dingding"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

//func newWithSeconds() *cron.Cron {
//	secondParser := cron.NewParser(cron.Second | cron.Minute |
//		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
//	return cron.New(cron.WithParser(secondParser), cron.WithChain())
//}
func InitCorn() (err error) {
	var Gcontab = cron.New(cron.WithSeconds()) //精确到秒
	global.GLOAB_CORN = Gcontab
	global.GLOAB_CORN.Start()
	//重启定时任务
	if err = Reboot(); err != nil {
		zap.L().Error(fmt.Sprintf("重启定时任务失败:%v\n", err))
	}
	zap.L().Debug("重启定时任务成功...")
	//重启考勤
	err = AttendanceByRobot()
	if err != nil {
		zap.L().Error("AttendanceByRobot init fail...")
	}
	zap.L().Debug("AttendanceByRobot init success...")
	err = RegularlySendCourses()
	if err != nil {
		return err
	}
	//发送爬取力扣的题目数
	err = SendLeetCode()
	if err != nil {
		zap.L().Error("SendLeetCode init fail...")
	}
	//重启考勤周报
	err = dingding.AttendWeeklyNewsPaper()
	if err != nil {
		zap.L().Error("AttendWeeklyNewsPaper init fail...")
	}
	return err
}
