package dingding

import (
	"context"
	"ding/global"
	"ding/model/common/response"
	"ding/model/params"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"runtime"
	"sort"
	"strconv"
	"time"
)

//获取到考勤部门
func GetAttendGroup() (err error, groupList []DingAttendGroup) {
	err = global.GLOAB_DB.Find(&groupList).Error
	return
}

//考勤周报
func AttendWeeklyNewsPaper() (err error) {
	//获取考勤组列表
	err, groupList := GetAttendGroup()
	if err != nil {
		zap.L().Error("获取考勤组列表失败", zap.Error(err))
		return
	}
	for _, group := range groupList {
		//根据考勤组获取成员信息
		if group.IsRobotAttendance == false {
			zap.L().Warn(fmt.Sprintf("考勤组：%v 开启未机器人考勤", group.GroupName))
			continue
		}
		p := &params.ParamAllDepartAttendByRobot{GroupId: group.GroupId}
		err := CountNum(group, p)
		if err != nil {
			zap.L().Error("打卡信息获取失败", zap.Error(err))
			continue
		}
	}
	return err
}

//开始统计周报
func CountNum(group DingAttendGroup, p *params.ParamAllDepartAttendByRobot) (err error) {
	//使用定时器,定时什么时候发送,每周的周日发
	spec := ""
	if runtime.GOOS == "windows" {
		spec = "00 02,11,11 9,14,20 * * ?"
	} else if runtime.GOOS == "linux" {
		spec = "0 15 10 ? * SAT"
	}
	zap.L().Info(spec)
	//获取当前的年份
	year := time.Now().Year()
	//获取当前是第几周,这个存到redis中方便去维护
	//维护周
	//判断当前年份，如果当前年份大于等于9则判定为下学期
	//默认为上学期
	upordown := 1
	month := time.Now().Month()
	if month >= 9 {
		upordown = 2
	}
	//获取当前周数
	err, week := GetWeek()
	//开启定时任务
	task := func() {
		//获取这个组织里面的成员信息
		token, _ := (&DingToken{}).GetAccessToken()
		g := DingAttendGroup{GroupId: p.GroupId, DingToken: DingToken{Token: token}}
		depts, err := g.GetGroupDeptNumber()
		if err != nil {
			zap.L().Error("获取考勤组部门成员(已经筛掉了不参与考勤的个人)失败", zap.Error(err))
			return
		}
		//获取考勤数据
		for _, dingUsers := range depts {
			num := make(map[string]int64)
			for _, user := range dingUsers {
				//拿到redis里面存的考勤信息
				WeekSignNum, err := user.GetWeekSignNum(year, upordown, week)
				if err != nil {
					zap.L().Error("统计失败", zap.Error(err))
					continue
				}
				num[user.UserId] = WeekSignNum
			}
			//查完一个部门后进行排序，然后发给用户
			sortnum := SortResult(num)
			//定义一个排名用于记录该用户排名
			ranking := 1
			for _, n := range sortnum {
				var ids []string
				ids[0] = n.UserId
				//封装要发送的消息
				message := EditMessage(ranking, n.WeekSignNum)
				//将考勤数据发给该人
				p := &ParamChat{
					RobotCode: "dingepndjqy7etanalhi",
					UserIds:   ids,
					MsgKey:    "sampleText",
					MsgParam:  message,
				}
				err = (&DingRobot{}).ChatSendMessage(p)
				ranking++
			}
		}
		//每次执行完后就去更新周数
		week += 1
		err = Week(week)
		if err != nil {
			if err != nil {
				zap.L().Error("周数设置失败，请尽快联系管理员", zap.Error(err))
			}
		}
	}
	_, err = global.GLOAB_CORN.AddFunc(spec, task)
	if err != nil {
		zap.L().Error("启动周报推送定时任务失败", zap.Error(err))
		return
	}
	return err
}

//编辑发送的消息
func EditMessage(ranking int, num int64) (message string) {
	message = "本周打卡记录：" + "\n" +
		"打卡次数：" + strconv.Itoa(int(num)) + "\n" +
		"打卡异常次数：" + strconv.Itoa(int(18-int(num))) + "\n" +
		"你在你部门的排名为：" + strconv.Itoa(ranking)
	return
}

type WeeklyNewPaper struct {
	UserId      string
	WeekSignNum int64
}

//排序
func SortResult(num map[string]int64) (WeeklyNewPapers []WeeklyNewPaper) {
	for k, v := range num {
		WeeklyNewPapers = append(WeeklyNewPapers, WeeklyNewPaper{k, v})
	}
	//这个地方也可以使用快排排序
	sort.Slice(WeeklyNewPapers, func(i, j int) bool {
		return WeeklyNewPapers[i].WeekSignNum > WeeklyNewPapers[j].WeekSignNum
	})
	return WeeklyNewPapers
	return
}

//向redis里面添加周
func Week(week int) (err error) {
	//设置一年的过期时间
	err = global.GLOBAL_REDIS.SetNX(context.Background(), "Week", week, 3600*time.Second*24*365).Err()
	return
}

//获取当前周
func GetWeek() (err error, week int) {
	str, err := global.GLOBAL_REDIS.Get(context.Background(), "Week").Result()
	week, err = strconv.Atoi(str)
	return
}

func ResetWeek(c *gin.Context) {
	err := global.GLOBAL_REDIS.SetNX(context.Background(), "Week", 1, 3600*time.Second*24*365).Err()
	if err != nil {
		zap.L().Error("重置失败", zap.Error(err))
		response.FailWithMessage("重置失败，请尽快联系管理员解决", c)
		return
	}
}
