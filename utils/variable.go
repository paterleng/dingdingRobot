package utils

import "time"

var (
	Time                  = 1
	MorningStartTime, _   = time.Parse("15:04:05", "08:00:00") //上午发送考勤开始时间
	MorningEndTime, _     = time.Parse("15:04:05", "08:30:00") //上午发送考勤开始时间
	AfternoonStartTime, _ = time.Parse("15:04:05", "14:30:00") //下午发送考勤开始时间
	AfternoonEndtTime, _  = time.Parse("15:04:05", "15:00:00") //下午发送考勤开始时间
	EveningStartTime, _   = time.Parse("15:04:05", "19:30:00") //晚上发送考勤开始时间
	EveningEndTime, _     = time.Parse("15:04:05", "20:30:00") //晚上发送考勤开始时间
	//推送各部门第一位打卡成员
	AttendanceMorningTime, _   = time.Parse("15:04:05", "11:30:00") //上午发送考勤开始时间
	AttendanceAfternoonTime, _ = time.Parse("15:04:05", "17:30:00") //上午发送考勤开始时间
	AttendanceEveningTime, _   = time.Parse("15:04:05", "23:59:59") //上午发送考勤开始时间
	StartHour                  = 22
	StartMin                   = 00
	RemindHour                 = 22
	RemindMin                  = 20
	EndHour                    = 22
	EndMin                     = 35
)
