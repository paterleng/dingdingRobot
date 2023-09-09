package dingding

import (
	"context"
	"ding/global"
	redis2 "ding/initialize/redis"
	"ding/model/common"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Task struct {
	gorm.Model
	TaskID            string              `json:"task_id" `  //cron第三方定时库给的id
	TaskName          string              `json:"task_name"` //任务名字
	UserId            string              `json:"user_id"`   // 任务所属userID
	UserName          string              `json:"user_name"` //任务所属用户
	RobotId           string              `json:"robot_id"`  //任务属于机器人
	Secret            string              `json:"secret"`
	RobotName         string              `json:"robot_name"`
	DetailTimeForUser string              `json:"detail_time_for_user"` //这个给用户看
	Spec              string              `json:"spec"`                 //这个是cron第三方的定时规则
	FrontRepeatTime   string              `json:"front_repeat_time"`    // 这个是前端传来的原始数据
	FrontDetailTime   string              `json:"front_detail_time"`
	MsgText           *common.MsgText     `json:"msg_text"`
	MsgLink           *common.MsgLink     `json:"msg_link"`
	MsgMarkDown       *common.MsgMarkDown `json:"msg_mark_down"`
	NextTime          time.Time           `json:"next_time"`
}

//编辑定时任务内容
type EditTaskContentParam struct {
	ID      uint   `json:"id"`
	TaskID  string `json:"task_id"`
	Content string `json:"content"`
}

func (t *Task) InsertTask() (err error) {
	//我先找一下数据库中与该任务相同的id号码，如果相同的话，说明数据库中有死掉的任务，需要加上软删除
	Dtask := []Task{}
	//找到所有的死任务，进行软删除
	global.GLOAB_DB.Where("task_id = ?", t.TaskID).Find(&Dtask)
	for i := 0; i < len(Dtask); i++ {
		err := global.GLOAB_DB.Delete(&Dtask[i]).Error
		if err != nil {
			zap.L().Error("软删除死任务失败", zap.Error(err))
		}
	}
	//然后再创建任务
	err = global.GLOAB_DB.Create(&t).Error
	return
}
func (t *Task) GetAllActiveTask() (tasks []Task, err error) {
	//先删除所有的任务，然后再重新加载一遍
	activeTasksKeys, err := global.GLOBAL_REDIS.Keys(context.Background(), fmt.Sprintf("%s*", redis2.Perfix+redis2.ActiveTask)).Result()
	if err != nil {
		zap.L().Error("从redis中获取旧的活跃任务的key失败", zap.Error(err))
		return
	}
	//删除所有的key
	global.GLOBAL_REDIS.Del(context.Background(), activeTasksKeys...)
	//拿到所有的任务的id
	//entries := global.GLOAB_CORN.Entries()
	//拿到所有任务的id
	//var entriesInt = make([]int, len(entries))
	//for index, value := range entries {
	//	entriesInt[index] = int(value.ID)
	//}
	// 根据id查询数据库，拿到详细的任务信息，存放到redis中
	global.GLOAB_DB.Model(&tasks).Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").Where("deleted_at is null").Find(&tasks)
	//查询所有的在线任务
	//把找到的数据存储到redis中 ，现在先写成手动获取
	//应该是存放在一个集合里面，集合里面存放着此条任务的所有信息，以id作为标识
	//哈希特别适合存储对象，所以我们用哈希来存储

	for _, task := range tasks {
		taskValue, err := json.Marshal(task) //把对象序列化成为一个json字符串
		if err != nil {
			zap.L().Info("定时任务序列化失败", zap.Error(err))
			continue
		}
		err = global.GLOBAL_REDIS.Set(context.Background(), redis2.GetTaskKey(task.TaskID), string(taskValue), 0).Err()
		if err != nil {
			zap.L().Error(fmt.Sprintf("从mysql获取所有活跃任务存入redis失败，失败任务id：%s，任务名：%s,执行人：%s,对应机器人：%s", task.TaskID, task.TaskName, task.UserName, task.RobotName), zap.Error(err))
			continue
		}
	}
	return
}
