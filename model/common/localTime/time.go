package localTime

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type MySelfTime struct {
	TimeStamp   int64
	Format      string // 完整的时间字符串
	Time        time.Time
	Duration    int //上午 下午 晚上 1 2 3
	ClassNumber int //当前是第几课节
	Week        int //周几
}

//根据考勤组判断当前时间（时间戳，字符串，time.Time,上午还是下午（根据考勤组规则制定））
func (t *MySelfTime) GetCurTime(commutingTime map[string][]string) (T MySelfTime, err error) {
	time.Sleep(1 * time.Second)
	zap.L().Info("进入到了自己封装的时间结构体中")
	timeStamp := time.Now()
	//获取到时间戳
	T.TimeStamp = timeStamp.UnixMilli()
	//time.Time转成字符串
	StringCurTime := timeStamp.Format("2006-01-02 15:04:05")
	T.Format = StringCurTime
	//字符串转成时间格式
	CurTime, _ := time.Parse("2006-01-02 15:04:05", StringCurTime)
	T.Time = CurTime
	zap.L().Info(fmt.Sprintf("当前时间的时间戳：%v,time.Time：%v,字符串格式：%s", T.TimeStamp, T.Time, T.Format))
	if commutingTime == nil || len(commutingTime) == 0 {
		zap.L().Info("commutingTime为空")
		AfternoonStart, _ := time.Parse("2006-01-02 15:04:05", StringCurTime[0:10]+" 12:00:00")
		EveningStart, _ := time.Parse("2006-01-02 15:04:05", StringCurTime[0:10]+" 19:00:00")
		zap.L().Info(fmt.Sprintf("上午下午时间分界点为：%s", AfternoonStart))
		zap.L().Info(fmt.Sprintf("下午晚上时间分界点为：%s", EveningStart))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.Before(AfternoonStart) 的值为:%v", CurTime, CurTime.Before(AfternoonStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(AfternoonStart) && CurTime.Before(EveningStart):%v", CurTime, CurTime.After(AfternoonStart) && CurTime.Before(EveningStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(EveningStart) 的值为:%v", CurTime, CurTime.After(EveningStart)))
		if CurTime.Before(AfternoonStart) {
			T.Duration = 1
		} else if CurTime.After(AfternoonStart) && CurTime.Before(EveningStart) {
			T.Duration = 2
		} else if CurTime.After(EveningStart) {
			T.Duration = 3
		}
		zap.L().Info(fmt.Sprintf("T.Duration = %v", T.Duration))
		if T.Duration == 0 {
			zap.L().Info("直接用时间对比，判断现在是上午还是下午失败，我们使用时间字符串，截取到小时，来判断")
			atoi, _ := strconv.Atoi(strings.Split(strings.Split(T.Format, " ")[1], ":")[0])
			zap.L().Info(fmt.Sprintf("截取到的小时为%v", atoi))
			if atoi < 12 {
				zap.L().Info("小于12，是上午")
				T.Duration = 1
			} else if atoi > 12 && atoi < 18 {
				zap.L().Info("大于12&&小于18，是下午")
				T.Duration = 2
			} else if atoi > 18 {
				zap.L().Info("大于18，是晚上")
				T.Duration = 3
			}
		}
		return
	}

	OnDuty := commutingTime["OnDuty"]
	//OffDuty := commutingTime["OffDuty"]
	//MorningStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[0])
	//MorningEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[0])
	if len(OnDuty) == 3 {
		AfternoonStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[1])
		//AfternoonEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[1])
		EveningStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[2])
		//EveningEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[2])
		zap.L().Info(fmt.Sprintf("上午下午时间分界点为：%s", AfternoonStart))
		zap.L().Info(fmt.Sprintf("下午晚上时间分界点为：%s", EveningStart))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.Before(AfternoonStart) 的值为:%v", CurTime, CurTime.Before(AfternoonStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(AfternoonStart) && CurTime.Before(EveningStart):%v", CurTime, CurTime.After(AfternoonStart) && CurTime.Before(EveningStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(EveningStart) 的值为:%v", CurTime, CurTime.After(EveningStart)))
		if CurTime.Before(AfternoonStart) {
			T.Duration = 1 //上午
		} else if CurTime.After(AfternoonStart) && CurTime.Before(EveningStart) {
			T.Duration = 2 //下午
		} else if CurTime.After(EveningStart) {
			T.Duration = 3
		}
		if T.Duration == 0 {
			zap.L().Info("直接用时间对比，判断现在是上午还是下午失败，我们使用时间字符串，截取到小时，来判断")
			atoi, _ := strconv.Atoi(strings.Split(strings.Split(T.Format, " ")[1], ":")[0])
			zap.L().Info(fmt.Sprintf("截取到的小时为%v", atoi))
			if atoi < 12 {
				zap.L().Info("小于12，是上午")
				T.Duration = 1
			} else if atoi > 12 && atoi < 18 {
				zap.L().Info("大于12&&小于18，是下午")
				T.Duration = 2
			} else if atoi > 18 {
				zap.L().Info("大于18，是晚上")
				T.Duration = 3
			}
		}
		T.ClassNumber = 1 //直接判定成第一节课
	} else if len(OnDuty) == 5 {
		zap.L().Info("进入第二节课考勤判定")
		//上午第二节课开始
		MorningSecondClassStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[1])
		//下午第二节课开始
		AfternoonSecondClassStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[3])
		AfternoonStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[2])

		//AfternoonEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[1])

		EveningStart, _ := time.Parse("2006-01-02 15:04:05", OnDuty[4]) //晚上上班
		//EveningEnd, _ := time.Parse("2006-01-02 15:04:05", OffDuty[2])
		zap.L().Info(fmt.Sprintf("上午下午时间分界点为：%s", AfternoonStart))
		zap.L().Info(fmt.Sprintf("下午晚上时间分界点为：%s", EveningStart))
		zap.L().Info(fmt.Sprintf("上午第一节课和第二节课的分界点为：%v", MorningSecondClassStart))
		zap.L().Info(fmt.Sprintf("下午第一节课和第二节课的分界点为：%v", AfternoonSecondClassStart))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.Before(AfternoonStart) 的值为:%v", CurTime, CurTime.Before(AfternoonStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(AfternoonStart) && CurTime.Before(EveningStart):%v", CurTime, CurTime.After(AfternoonStart) && CurTime.Before(EveningStart)))
		zap.L().Info(fmt.Sprintf("当前时间为：%v，CurTime.After(EveningStart) 的值为:%v", CurTime, CurTime.After(EveningStart)))
		T.ClassNumber = 1
		if CurTime.Before(AfternoonStart) {
			zap.L().Info("现在是上午时间")
			T.Duration = 1
			zap.L().Info(fmt.Sprintf("CurTime.After(MorningSecondClassStart)为 %v", CurTime.After(MorningSecondClassStart)))
			if CurTime.After(MorningSecondClassStart) {
				zap.L().Info("CurTime.After(MorningSecondClassStart) 为true,现在是上午第二节课")
				T.ClassNumber = 2
			}
		} else if CurTime.After(AfternoonStart) && CurTime.Before(EveningStart) {
			zap.L().Info("成功判定当前为下午，T.Duration = 2")
			T.Duration = 2
			zap.L().Info(fmt.Sprintf("CurTime为%v,AfternoonSecondClassStart为%v,CurTime.After(AfternoonSecondClassStart)的值为%v", CurTime, AfternoonSecondClassStart, CurTime.After(AfternoonSecondClassStart)))
			if CurTime.After(AfternoonSecondClassStart) {
				zap.L().Info("成功判定当前是下午第二节课")
				T.ClassNumber = 2
			}
		} else if CurTime.After(EveningStart) {
			T.Duration = 3
		}
		if T.Duration == 0 {
			zap.L().Info("直接用时间对比，判断现在是上午还是下午失败，我们使用时间字符串，截取到小时，来判断")
			hour, _ := strconv.Atoi(strings.Split(strings.Split(T.Format, " ")[1], ":")[0])
			zap.L().Info(fmt.Sprintf("截取到的小时为%v", hour))
			if hour < 12 {
				zap.L().Info("小于12，是上午")
				T.Duration = 1

				atoi, _ := strconv.Atoi(OnDuty[2])
				if hour > atoi {
					T.ClassNumber = 2
				}
			} else if hour > 12 && hour < 18 {
				zap.L().Info("大于12&&小于18，是下午")
				T.Duration = 2
				atoi, _ := strconv.Atoi(OnDuty[4])
				if hour > atoi {
					T.ClassNumber = 2
				}
			} else if hour > 18 {
				zap.L().Info("大于18，是晚上")
				T.Duration = 3
			}
		}
	}

	if T.Duration == 0 {
		zap.L().Info("直接用时间对比，判断现在是上午还是下午失败，我们使用时间字符串，截取到小时，来判断")
		atoi, _ := strconv.Atoi(strings.Split(strings.Split(T.Format, " ")[1], ":")[0])
		zap.L().Info(fmt.Sprintf("截取到的小时为%v", atoi))
		if atoi < 12 {
			zap.L().Info("小于12，是上午")
			T.Duration = 1
		} else if atoi > 12 && atoi < 18 {
			zap.L().Info("大于12&&小于18，是下午")
			T.Duration = 2
		} else if atoi > 18 {
			zap.L().Info("大于18，是晚上")
			T.Duration = 3
		}
	}
	zap.L().Info(fmt.Sprintf("T = %v", T))
	return
}

func (t *MySelfTime) StringToStamp(s string) (int64, error) {
	if s == "" && t.Format != "" {
		timeT, err := time.ParseInLocation("2006-01-02 15:04:05", t.Format, time.Local)
		if err != nil {
			return 0, errors.New("时间转化成时间戳失败")
		}
		return timeT.Unix() * 1000, err
	}
	timeT, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		return 0, errors.New("时间转化成时间戳失败")
	}
	return timeT.Unix() * 1000, err
}
func (t *MySelfTime) StampToString(s int64) string {
	//如果传递参数是0，且t.TimeStamp != 0 ,默认获取当前时间的时间戳
	if s == 0 && t.TimeStamp != 0 {
		return time.Unix(t.TimeStamp/1000, 0).Format("2006-01-02 15:04:05")
	}
	return time.Unix(s/1000, 0).Format("2006-01-02 15:04:05")

}
func (t *MySelfTime) GetWeek(T *time.Time) string {
	if T != nil {
		return T.Weekday().String()
	}

	return t.Time.Weekday().String()
}
