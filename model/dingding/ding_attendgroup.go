package dingding

import (
	"context"
	"crypto/tls"
	"ding/global"
	"ding/initialize/redis"
	"ding/model/classCourse"
	"ding/model/common"
	"ding/model/common/localTime"
	"ding/model/common/request"
	"ding/model/params"
	"ding/model/params/ding"
	"ding/utils"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	redisZ "github.com/go-redis/redis/v8"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DingAttendGroup struct {
	CreatedAt     time.Time      `gorm:"column:create_at"`
	UpdatedAt     time.Time      `gorm:"column:update_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
	GroupId       int            `gorm:"primaryKey" json:"group_id"` //考勤组id
	GroupName     string         `json:"group_name"`                 //考勤组名称
	MemberCount   int            `json:"member_count"`               //参与考勤人员总数
	WorkDayList   []string       `gorm:"-" json:"work_day_list"`     //0表示休息,数组内的值，从左到右依次代表周日到周六，每日的排班情况。
	ClassesList   []string       `gorm:"-" json:"classes_list"`      // 一周的班次时间展示列表
	SelectedClass []struct {
		Setting struct {
			PermitLateMinutes int `json:"permit_late_minutes"` //允许迟到时长
		} `gorm:"-" json:"setting"`
		Sections []struct {
			Times []struct {
				CheckTime string `json:"check_time"` //打卡时间
				CheckType string `json:"check_type"` //打卡类型
			} `gorm:"-" json:"times"`
		} `gorm:"-" json:"sections"`
	} `gorm:"-" json:"selected_class"`
	DingToken         `gorm:"-"`
	IsRobotAttendance bool       `json:"is_robot_attendance"` //该考勤组是否开启机器人查考勤 （相当于是总开关）
	RobotAttendTaskID int        `json:"robot_attend_task_id"`
	IsSendFirstPerson int        `json:"is_send_first_person"` //该考勤组是否开启推送每个部门第一位打卡人员 （总开关）
	IsInSchool        bool       `json:"is_in_school"`         //是否在学校，如果在学校，开启判断是否有课
	IsReady           int        `json:"is_ready"`             //是否预备
	ReadyTime         int        `json:"ready_time"`           //如果预备了，提前几分钟
	NextTime          string     `json:"next_time"`            //下次执行时间
	IsSecondClass     int        `json:"is_second_class"`      //是否开启第二节课考勤
	RestTimes         []RestTime `json:"rest_times" gorm:"foreignKey:AttendGroupID;references:group_id"`
}
type RestTime struct {
	gorm.Model    // 1 2 2 0 2 1
	WeekDay       int
	MAE           int // 0 1 2
	AttendGroupID int
}
type DingAttendanceGroupMemberList struct {
	AtcFlag  string `json:"atc_flag"`
	Type     string `json:"type"`
	MemberID string `json:"member_id"`
}

//通过考勤组id获取休息时间
//func (DingAttendGroup *DingAttendGroup) GetRestTimes(weekDay int) (RestTimes []RestTime) {
//	//err := global.GLOAB_DB.Model(&DingAttendGroup).Preload("RestTimes").Find(&RestTimes).Error
//	//if err != nil {
//	//	zap.L().Error("")
//	//}
//
//	//err := global.GLOAB_DB.Model(&DingAttendGroup).Association("RestTimes").Find(&RestTimes).Error
//	//if err != nil {
//	//	zap.L().Error("")
//	//}
//	//return RestTimes
//
//	err := global.GLOAB_DB.Preload("RestTimes").First(DingAttendGroup)
//	if err != nil {
//		zap.L().Error("")
//	}
//	RestTimes = make([]RestTime, len(DingAttendGroup.RestTimes))
//	RestTimes = DingAttendGroup.RestTimes
//	return RestTimes
//}
func (DingAttendGroup *DingAttendGroup) BeforeCreate(tx *gorm.DB) (err error) {
	DingAttendGroup.CreatedAt = time.Now()
	return
}

func (DingAttendGroup *DingAttendGroup) BeforeUpdate(tx *gorm.DB) (err error) {
	DingAttendGroup.UpdatedAt = time.Now()
	return
}

// 批量获取考勤组
func (a *DingAttendGroup) GetAttendancesGroups(offset int, size int) (groups []DingAttendGroup, err error) {
	if !a.DingToken.IsLegal() {
		return nil, errors.New("token不合法")
	}
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/attendance/getsimplegroups?access_token=" + a.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		Offset int
		Size   int
	}{
		Offset: offset,
		Size:   size,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			Groups []DingAttendGroup `json:"groups"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	groups = r.Result.Groups
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_name", "member_count"}),
	}).Create(&groups).Error
	return groups, nil
}

// 获取一天的上下班时间
// map["OnDuty"] map["OffDuty"]
func (a *DingAttendGroup) GetCommutingTime() (FromTo map[string][]string, err error) {
	FromTo = make(map[string][]string, 2)
	timeNowYMD := time.Now().Format("2006-01-02")
	attendancesGroupsDetail, err := a.GetAttendancesGroupById()
	if err != nil {
		return
	}
	Sections := attendancesGroupsDetail.SelectedClass[0].Sections //上午中午下午三个模块
	OnDutyTime := make([]string, 0)
	OffDutyTime := make([]string, 0)
	for _, section := range Sections {
		for _, time := range section.Times {
			l := len(time.CheckTime)
			b := []byte(time.CheckTime[l-8:])
			if time.CheckType == "OnDuty" {
				b[4] = utils.Delay
				OnDutyTime = append(OnDutyTime, timeNowYMD+" "+string(b))
			} else {
				//OffDutyTime = append(OffDutyTime, timeNowYMD+" "+time.CheckTime[l-8:])
				OffDutyTime = append(OffDutyTime, timeNowYMD+" "+string(b))
			}
			FromTo["OnDuty"] = OnDutyTime
			FromTo["OffDuty"] = OffDutyTime
		}
	}
	return
}

//用于打卡提醒获取上班时间段
func (a *DingAttendGroup) GetCommutingTime1() (FromTo map[string][]string, err error) {
	FromTo = make(map[string][]string, 2)
	timeNowYMD := time.Now().Format("2006-01-02")
	attendancesGroupsDetail, err := a.GetAttendancesGroupById()
	if err != nil {
		return
	}
	Sections := attendancesGroupsDetail.SelectedClass[0].Sections //上午中午下午三个模块
	OnDutyTime := make([]string, 0)
	OffDutyTime := make([]string, 0)
	for _, section := range Sections {
		for _, time := range section.Times {
			l := len(time.CheckTime)
			b := []byte(time.CheckTime[l-8:])
			str := string(b)
			if time.CheckType == "OnDuty" {
				s := strings.Split(str, ":")
				h, _ := strconv.Atoi(s[0])
				m, _ := strconv.Atoi(s[1])
				//先转化成分钟
				i2 := h*60 + m
				m = (i2 - 5) % 60
				h = (i2 - 5) / 60
				minute := strconv.Itoa(m)
				hours := strconv.Itoa(h)
				hour := hours + ":"
				min := minute + ":"
				second := "00"
				times := hour + min + second
				OnDutyTime = append(OnDutyTime, timeNowYMD+" "+times)
			} else {
				//OffDutyTime = append(OffDutyTime, timeNowYMD+" "+time.CheckTime[l-8:])
				OffDutyTime = append(OffDutyTime, timeNowYMD+" "+string(b))
			}
			FromTo["OnDuty"] = OnDutyTime
			FromTo["OffDuty"] = OffDutyTime
		}
	}
	return
}

func (a *DingAttendGroup) GetWorkDayList() ([]string, error) {
	attendancesGroupsDetail, err := a.GetAttendancesGroupById()
	if err != nil {
		return attendancesGroupsDetail.ClassesList, err
	}

	return attendancesGroupsDetail.ClassesList, err
}

// 根据id获取详细的考勤组
func (a *DingAttendGroup) GetAttendancesGroupById() (group DingAttendGroup, err error) {
	groups, err := a.GetAttendancesGroups(0, 50)
	if err != nil {
		return
	}
	for _, attendGroup := range groups {
		if strconv.Itoa(attendGroup.GroupId) == strconv.Itoa(a.GroupId) {
			group = attendGroup
			break
		}
	}
	return
}

// 获取考勤组中的部门成员，已经筛掉了不参与考勤的人员
func (a *DingAttendGroup) GetGroupDeptNumber() (DeptUsers map[string][]DingUser, err error) {
	DeptUsers = make(map[string][]DingUser)
	result, err := a.GetAttendancesGroupMemberList("413550622937553255")
	//存储不参与考勤人员，键是用户id，值是用户名
	NotAttendanceUserIdListMap := make(map[string]string)
	DeptAllUserList := make([]DingUser, 0)
	for _, Member := range result {
		if Member.Type == "0" && Member.AtcFlag == "1" { //单个人且不参与考勤
			u := DingUser{
				UserId:    Member.MemberID,
				DingToken: a.DingToken,
			}
			NotAttendanceUser, err := u.GetUserDetailByUserId()
			if err != nil {
				zap.L().Error(fmt.Sprintf("找不到单个人且不参与考勤 的个人信息，跳过%v", u))
				continue
			}
			NotAttendanceUserIdListMap[Member.MemberID] = NotAttendanceUser.Name
		}
	}
	for _, Member := range result {
		DeptAttendanceUserList := make([]DingUser, 0)
		if Member.Type == "1" && Member.AtcFlag == "0" { //部门且参与考勤
			deptId, _ := strconv.Atoi(Member.MemberID)
			d := DingDept{DeptId: deptId, DingToken: DingToken{Token: a.DingToken.Token}}
			DeptAllUserList, _, err = d.GetUserListByDepartmentID(0, 100)
			if err != nil {
				zap.L().Error(fmt.Sprintf("通过部门id:%v获取部门用户列表失败", deptId), zap.Error(err))
				continue
			}
			for _, value := range DeptAllUserList {
				if _, ok := NotAttendanceUserIdListMap[value.UserId]; ok {
					continue
				}
				DeptAttendanceUserList = append(DeptAttendanceUserList, value)
			}
			DeptUsers[Member.MemberID] = DeptAttendanceUserList
		}
	}
	return
}

// 获取考勤组人员（部门id和成员id）https://open.dingtalk.com/document/isvapp-server/batch-query-of-attendance-group-members
func (a *DingAttendGroup) GetAttendancesGroupMemberList(OpUserID string) (R []DingAttendanceGroupMemberList, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/attendance/group/member/list?access_token=" + a.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		OpUserID string `json:"op_user_id"`
		GroupID  int    `json:"group_id"`
	}{
		OpUserID: OpUserID,
		GroupID:  a.GroupId,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			DingAttendanceGroupMemberList []DingAttendanceGroupMemberList `json:"result"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	R = r.Result.DingAttendanceGroupMemberList
	return R, nil
}

// 通过部门id获取部门所有成员user_id（非详细信息） https://open.dingtalk.com/document/isvapp-server/query-the-list-of-department-userids
func (a *DingAttendGroup) GetUserListByDepartmentID(token string, deptId, cursor, size int) (userList []DingUser, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/list?access_token=" + token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
		Cursor int `json:"cursor"`
		Size   int `json:"size"`
	}{
		DeptID: deptId,
		Cursor: cursor,
		Size:   size,
	}
	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		DingResponseCommon
		Result struct {
			HasMore bool       `json:"has_more"`
			List    []DingUser `json:"list"`
		} `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	return r.Result.List, nil
}

// 更新数据库考勤组
func (a *DingAttendGroup) UpdateAttendGroup(p *ding.ParamUpdateUpdateAttendanceGroup) (err error) {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		var old DingAttendGroup
		err = tx.First(&old, p.GroupId).Error
		if err != nil {
			return err
		}
		AttendGroup := &DingAttendGroup{GroupId: p.GroupId, IsSendFirstPerson: p.IsSendFirstPerson, IsRobotAttendance: p.IsRobotAttendance, IsReady: p.IsReady, ReadyTime: p.ReadyTime}
		//err = tx.Updates(AttendGroup).Error
		//if err != nil {
		//	return err
		//}
		if old.IsRobotAttendance == false && AttendGroup.IsRobotAttendance == true {
			zap.L().Error("更新考勤组开启定时任务")
			//开启定时任务
			P := &params.ParamAllDepartAttendByRobot{
				GroupId: p.GroupId,
			}
			_, taskID, err := a.AllDepartAttendByRobot(P)
			if err != nil {
				zap.L().Error("开启定时任务AllDepartAttendByRobot()失败", zap.Error(err))
				return err
			}
			AttendGroup.RobotAttendTaskID = int(taskID)
			AttendGroup.IsRobotAttendance = true
			err = tx.Model(&AttendGroup).Updates(AttendGroup).Error
			if err != nil {
				zap.L().Error("mysql更新考勤组定时任务task_id失败")
			}
			zap.L().Info(fmt.Sprintf("开启考勤组考勤定时任务成功！定时任务id为%d", taskID))
			return err
		} else if old.IsRobotAttendance == true && AttendGroup.IsRobotAttendance == false {
			zap.L().Error("更新考勤组关闭定时任务")
			AttendGroup.RobotAttendTaskID = -1
			AttendGroup.IsRobotAttendance = false
			err = tx.Updates(AttendGroup).Error
			if err != nil {
				zap.L().Error("更新考勤组定时任务id为-1失败", zap.Error(err))
			}
			//updates不会更新零值，所以我们使用update单独更新一下
			err = tx.Model(&AttendGroup).Update("is_robot_attendance", false).Error
			if err != nil {
				return err
			}
			zap.L().Info(fmt.Sprintf("关闭cron定时任务，定时任务id为：%v", old.RobotAttendTaskID))
			global.GLOAB_CORN.Remove(cron.EntryID(old.RobotAttendTaskID))
			zap.L().Info("关闭考勤组考勤定时任务成功！")
		}
		return err
	})
}

// 获取数据库考勤组数据
func (a *DingAttendGroup) GetAttendanceGroupListFromMysql(info *request.PageInfo) (DingAttendGroupList []DingAttendGroup, err error) {
	err = global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		limit := info.PageSize
		offset := info.PageSize * (info.Page - 1)
		err = tx.Limit(limit).Offset(offset).Find(&DingAttendGroupList).Error
		if err != nil {
			return err
		}
		return err
	})
	return DingAttendGroupList, err
}

//判断是否在正确的执行时间
func CronHandle(spec string, curTime localTime.MySelfTime) (Ok bool) {
	s := strings.Split(spec, " ")
	min := strings.Split(s[1], ",")
	hour := strings.Split(s[2], ",")
	minHour := make([]string, 0)
	if len(min) != len(hour) && len(min) != 1 && len(hour) != 1 {
		zap.L().Error("spec不合法")
		return
	} else if len(min) > 1 && len(hour) > 1 && len(min) == len(hour) {
		zap.L().Info("使用spec一个表达式在多个不同的时间执行，很特殊的一种用法")
		//拼装时间
		for i := 0; i < len(min); i++ {
			minHour = append(minHour, hour[i]+":"+min[i])
		}
		curDate := curTime.Format[0:10]
		for i := 0; i < len(minHour); i++ {
			//拼装成完整的一天的该要运行的时间点
			minHour[i] = curDate + " " + minHour[i] + ":00"
		}
	}

	stamps := make([]int64, 0)
	for i := 0; i < len(minHour); i++ {
		//把时间转化成时间戳
		stamp, err := (&localTime.MySelfTime{}).StringToStamp(minHour[i])
		if err != nil {
			zap.L().Error("把一天的该要运行的时间点 字符串转化成int64时间戳失败", zap.Error(err))
			return
		}
		stamps = append(stamps, stamp)
	}
	OK := false
	//现在把需要运行的时间戳整了出来，不需要运行的直接跳过即可
	for i := 0; i < len(stamps); i++ {
		if curTime.TimeStamp > stamps[i]-1000*60 && curTime.TimeStamp < stamps[i]+1000*60 {
			OK = true
			break
		}
	}

	return OK
}

//处理日期
func DateHandle(curTime localTime.MySelfTime) (startWeek, week, CourseNumber int, err error) {

	m1 := map[string]int{"Sunday": 7, "Monday": 1, "Tuesday": 2, "Wednesday": 3, "Thursday": 4, "Friday": 5, "Saturday": 6}
	now := time.Now()
	weekEnglish := (&localTime.MySelfTime{}).GetWeek(&now)
	//周几
	week = m1[weekEnglish]
	//第几周
	startWeek, err = (&classCourse.Calendar{}).GetWeek()
	if err != nil {
		zap.L().Error("通过课表小程序获取当前第几周失败", zap.Error(err))
		//return 20, 0, 0, err
	}
	startWeek = 20
	//获取当前是第几节课
	if curTime.Duration == 1 {
		if curTime.ClassNumber == 1 {
			CourseNumber = 1
		} else if curTime.ClassNumber == 2 {
			CourseNumber = 2
		}
	} else if curTime.Duration == 2 {
		if curTime.ClassNumber == 1 {
			zap.L().Info("curT.Duration == 2 ,现在是下午，所以我们查第三课考勤")
			CourseNumber = 3
		} else if curTime.ClassNumber == 2 {
			zap.L().Info("curT.Duration == 2 ,现在是下午，所以我们查第四课考勤")
			CourseNumber = 4
		}

	} else if curTime.Duration == 3 {
		zap.L().Info("curT.Duration == 3 ,现在是晚上，所以我们查第五课考勤")
		CourseNumber = 5
	}
	return
}

// 该考勤组进行机器人考勤
func (a *DingAttendGroup) AllDepartAttendByRobot(p *params.ParamAllDepartAttendByRobot) (result map[string][]DingAttendance, taskID cron.EntryID, err error) {
	//判断一下是否需要需要课表小程序的数据
	token, err := (&DingToken{}).GetAccessToken()
	if err != nil || token == "" {
		zap.L().Error("从redis中取出token失败", zap.Error(err))
		return
	}
	g := DingAttendGroup{GroupId: p.GroupId, DingToken: DingToken{Token: token}}
	commutingTime, err := g.GetCommutingTime()
	if err != nil {
		zap.L().Error("根据考勤组获取上下班时间失败", zap.Error(err))
		return
	}
	//获取到上班时间
	OnDutyTimeList := commutingTime["OnDuty"]
	//获取到不考勤时间
	var restTime []RestTime
	err = global.GLOAB_DB.Where("attend_group_id", a.GroupId).Find(&restTime).Error
	if err != nil {
		zap.L().Error("根据考勤组获取休息时间失败", zap.Error(err))
		return
	}
	//把时间格式拼装处理一下，拼装成corn定时库spec定时规则能够使用的格式
	min := ""
	hour := ""
	for i := 0; i < len(OnDutyTimeList); i++ {
		s := strings.Split(strings.Split(OnDutyTimeList[i], " ")[1], ":")
		hour += s[0] + ","
		min += s[1] + ","
	}
	hour = hour[:len(hour)-1]
	min = min[:len(min)-1]
	spec := ""
	if runtime.GOOS == "windows" {
		spec = "00 02,11,31 9,14,20 * * ?"
	} else if runtime.GOOS == "linux" {
		spec = "00 " + min + " " + hour + " * * ?"
	}
	zap.L().Info(spec)
	task := func() {
		token, err = (&DingToken{}).GetAccessToken()
		g := DingAttendGroup{GroupId: p.GroupId, DingToken: DingToken{Token: token}}

		//a := DingAttendance{DingToken: DingToken{Token: token}}
		//获取一天上下班的时间
		commutingTimes, err := g.GetCommutingTime()
		if err != nil {
			zap.L().Error("根据考勤组id获取一天上下班失败失败", zap.Error(err))
			return
		}
		//获取上班时间、//获取下班时间
		OnDutyTime := commutingTimes["OnDuty"]
		OffDutyTime := commutingTimes["OffDuty"]
		zap.L().Info(fmt.Sprintf("上班时间：%v", OnDutyTime))
		zap.L().Info(fmt.Sprintf("下班时间：%v", OffDutyTime))
		//获取当前时间，curTime是自己封装的时间类型，有各种格式的时间
		curTime, err := (&localTime.MySelfTime{}).GetCurTime(commutingTimes)
		if err != nil {
			zap.L().Error("获取当前时间失败", zap.Error(err))
			return
		}
		//判断当前时间是否需要运行，我们使用的是cron定时器，corn定时器不支持一些不规则的定时，我们此处做一些判断，跳过一些不合法的时间
		ok := CronHandle(spec, curTime)
		if ok == false {
			zap.L().Info("当前时间cron执行，但是不是我们想要的时间，跳过执行")
			return
		}
		//获取考勤组部门成员，已经筛掉了不参与考勤的个人
		//注意一定要放在task里面，这样当纪检部更新了考勤组之后，每次加载人员都是最新的
		deptAttendanceUser, err := g.GetGroupDeptNumber()
		if err != nil {
			zap.L().Error("获取考勤组部门成员(已经筛掉了不参与考勤的个人)失败", zap.Error(err))
			return
		}
		zap.L().Info(fmt.Sprintf("考勤规则：%v，考勤人员详情：%v", spec, deptAttendanceUser))
		//判断该考勤组是否在校，在校的话，需要判定有课无课情况，如果不在校，则统一按照无课处理
		var isInSchool bool
		err = global.GLOAB_DB.Model(&DingAttendGroup{GroupId: p.GroupId}).Select("is_in_school").Scan(&isInSchool).Error
		if err != nil {
			zap.L().Error("通过考勤组判断是否在学校（加入课表小程序数据失败）", zap.Error(err))
			isInSchool = false
		}
		token, err := (&DingToken{}).GetAccessToken()
		if err != nil {
			zap.L().Error("从redis中取出token失败", zap.Error(err))
			return
		}
		Len := len(deptAttendanceUser)
		Count := 0
		startWeek, week, CourseNumber, err := DateHandle(curTime)
		//判断是不是freetime时间
		for _, rest := range restTime {
			if week == rest.WeekDay && curTime.Duration == rest.MAE {
				zap.L().Info("freetime跳过")
				//直接所有部门都不再发送了
				return
			}
		}
		for DeptId, _ := range deptAttendanceUser { //
			Count++
			atoi, _ := strconv.Atoi(DeptId)
			//DeptDetail, err := d.GetDeptByIDFromMysql()
			DeptDetail, err := (&DingDept{DingToken: DingToken{Token: token}, DeptId: atoi}).GetDeptByIDFromMysql()
			DeptDetail.DingToken.Token = token
			if err != nil {
				zap.L().Error(fmt.Sprintf("通过部门id：%s获取部门详情失败，继续执行下一轮循环", DeptId), zap.Error(err))
				continue
			}
			//todo 判断一下此部门是否开启推送考勤
			if DeptDetail.IsRobotAttendance == false || DeptDetail.RobotToken == "" {
				zap.L().Error(fmt.Sprintf("该部门:%s为开启考勤或者机器人robotToken:%s是空，跳过", DeptDetail.Name, DeptDetail.RobotToken))
				continue
			}
			zap.L().Info(fmt.Sprintf("该部门:%s开启考勤,机器人robotToken:%s", DeptDetail.Name, DeptDetail.RobotToken))
			result = make(map[string][]DingAttendance, 0)
			result["Normal"] = make([]DingAttendance, 0)
			result["Late"] = make([]DingAttendance, 0)
			result["Leave"] = make([]DingAttendance, 0)
			result["HasCourse"] = make([]DingAttendance, 0)
			//获取了一个部门所有参与考勤的用户id
			DeptAttendanceUserIdList := GetUserIdListByUserList(deptAttendanceUser[DeptId])
			//根据用户id获取用户打卡情况，同时返回了没有考勤数据的同学
			attendanceList, NotRecordUserIdList, err := DeptDetail.GetAttendanceData(DeptAttendanceUserIdList, curTime, OnDutyTime, OffDutyTime)
			if err != nil {
				zap.L().Error("根据部门用户id列表获取用户打卡情况失败", zap.Error(err))
			}
			//遍历考勤数据,有课的优先级高于考勤
			for _, attendance := range attendanceList {
				flag := false
				//查一下课表，有课且打卡的话，判定为有课
				if isInSchool {
					course, _ := classCourse.GetIsHasCourse(CourseNumber, startWeek, 0, []string{attendance.UserID}, week)
					for _, Byclass := range course {
						if Byclass.Userid == attendance.UserID {
							result["HasCourse"] = append(result["HasCourse"], attendance)
							flag = true
							break
						}
					}
				}

				if flag == false {
					if attendance.TimeResult == "Normal" {
						result["Normal"] = append(result["Normal"], attendance)
					}
				}
			}
			zap.L().Info(fmt.Sprintf("有考勤记录同学已经处理完成，接下来开始处理没有考勤数据的同学"))

			/*
				获取课表小程序有课的同学
				课表小程序有一个接口，可以获取到大家的有课无课情况，其中参数有
				当前周、高级筛选中的部门，我们找到部门中有课的同学，然后跳过即可
			*/
			//处理没有考勤记录的同学，看看其是否有课，map传递的引用类型
			if isInSchool {
				//调用课表小程序接口，判断没有考勤数据的人是否请假了
				//需要参数：当前周、周几、第几节课，NotRecordUserIdList
				//此处传递的两个参数 NotRecordUserIdList、result 都是引用类型，NotRecordUserIdList处理之后已经不含有课的成员了
				HasCourseHandle(NotRecordUserIdList, CourseNumber, startWeek, week, result)
			}

			//if (week == restTime[0].WeekDay && curTime.Duration == restTime[0].MAE) || (week == 2 && curTime.Duration == 1) || (week == 2 && curTime.Duration == 2) {
			//	zap.L().Info("freetime跳过")
			//	//直接所有部门都不再发送了
			//	return
			//}

			err, _ = LeaveLateHandle(NotRecordUserIdList, token, result, curTime)
			if err != nil {
				zap.L().Error("处理请假和迟到有误", zap.Error(err))
			}
			zap.L().Info("没有考勤数据的同学已经处理完成")
			//一个部门的考勤结束了，开始封装信息，发送考勤消息

			message := MessageHandle(curTime, DeptDetail, result)

			zap.L().Info("message编辑完成，开始封装发送信息参数")
			pSend := &ParamCronTask{
				MsgText: &common.MsgText{
					At: common.At{IsAtAll: false},
					Text: common.Text{
						Content: message,
					},
					Msgtype: "text",
				},
				RobotId: DeptDetail.RobotToken,
			}
			zap.L().Info(fmt.Sprintf("正在发送信息，信息参数为%v", pSend))
			err = (&DingRobot{RobotId: DeptDetail.RobotToken}).SendMessage(pSend)
			if err != nil {
				zap.L().Error(fmt.Sprintf("发送信息失败，信息参数为%v", pSend), zap.Error(err))
				continue
			}
			//在此处使用bitmap来实现存储功能
			err, week := GetWeek()
			err = BitMapHandle(result, curTime, startWeek, week)
			if err != nil {
				zap.L().Error("使用bitmap存储每个人的记录失败", zap.Error(err))
			}
			//将考勤数据发给部门负责人以及管理人员
			var userids []string
			global.GLOAB_DB.Table("user_dept").Where("is_responsible = ? and ding_dept_dept_id = ?", true, DeptId).Select("ding_user_user_id").Find(&userids)
			p := &ParamChat{
				RobotCode: "dingepndjqy7etanalhi",
				UserIds:   userids,
				MsgKey:    "sampleText",
				MsgParam:  message,
			}
			err = (&DingRobot{}).ChatSendMessage(p)
			if err != nil {
				zap.L().Error("发送至部门负责人失败", zap.Error(err))
			}
			//(DingRobot{}).CommonSingleChat()
			// 向各部门根据请假次数排序的集合中 设置key
			// 获取此次考勤该部门的请假次数
			zap.L().Info(fmt.Sprintf("部门：%v开始统计请假迟到信息到redis中", DeptDetail.Name))
			leaveCount := len(result["Leave"])
			// 该部门的总人数
			var deptNumbers float64 = float64(len(deptAttendanceUser[DeptId]))
			memberName := fmt.Sprintf("部门名称:%v", DeptDetail.Name)
			preAveScore := global.GLOBAL_REDIS.ZScore(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":", memberName).Val()
			zap.L().Info(fmt.Sprintf("取出该部门先前的平均请假次数：%v", preAveScore))
			// 该部门的平均请假率
			score, err := strconv.ParseFloat(fmt.Sprintf("%.6f", (preAveScore*deptNumbers+float64(leaveCount))/deptNumbers), 64)
			if err != nil {
				zap.L().Info("部门平均请假率转换失败")
				return
			}
			// 开启事务
			pipeline := global.GLOBAL_REDIS.TxPipeline()
			err = pipeline.ZAdd(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":", &redisZ.Z{
				// 根据平均请假次数排序
				Score:  score,
				Member: memberName,
			}).Err()

			// 记录此部门的请假总次数，拼装键，然后在键上面进行添加
			//这是普通的key value键值对
			err = pipeline.IncrBy(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":dept:"+DeptDetail.Name, int64(leaveCount)).Err()

			//登记部门里面每个人请假情况 （对zset进行操作）
			err = (&DeptDetail).CountFrequencyLeave(startWeek, result)
			if err != nil {
				zap.L().Info("CountFrequencyLeave失败", zap.Error(err))
			}

			// 提交事务
			_, err = pipeline.Exec(context.Background())
			// 命令执行失败，取消提交
			if err != nil {
				zap.L().Info("key:部门 value:请假总次数的集合键设置失败，部门id为：" + DeptId)
				pipeline.Discard()
				continue
			}
			pipeline.Close()
			//发送部门排行榜请假情况
			DeptDetail.SendFrequencyLeave(startWeek)

			// 以下是对迟到Zset的操作
			pipeline = global.GLOBAL_REDIS.TxPipeline()
			lateCount := len(result["Late"])
			preAveLateScore := global.GLOBAL_REDIS.ZScore(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":", memberName).Val()
			scoreAveLate, err := strconv.ParseFloat(fmt.Sprintf("%.6f", (preAveLateScore*deptNumbers+float64(lateCount))/deptNumbers), 64)

			// 对迟到Zset更新member的score
			pipeline.ZAdd(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":", &redisZ.Z{
				// 根据平均迟到次数排序
				Score:  scoreAveLate,
				Member: memberName,
			})
			err = (&DeptDetail).CountFrequencyLate(startWeek, result)
			if err != nil {

			}
			pipeline.IncrBy(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":dept:"+DeptDetail.Name, int64(lateCount))

			_, err = pipeline.Exec(context.Background())
			if err != nil {
				zap.L().Info("更新迟到有序集合事务提交失败")
				pipeline.Discard()
				continue
			}
			pipeline.Close()
			// 若是周日就发送各部门平均请假、迟到排行榜
			err = DeptDetail.SendFrequencyLate(startWeek) //部门个人请假排行榜
			//当遍历到map最后一个元素的时候，我们发送一下所有部门的请假和迟到情况
			if Count == Len {
				SundayAfternoonExec(startWeek)
			}
			zap.L().Info("信息发送成功" + message)
		}
		return
	}
	taskID, err = global.GLOAB_CORN.AddFunc(spec, task)
	if err != nil {
		zap.L().Error("启动机器人查考勤定时任务失败", zap.Error(err))
		return
	}
	nextTime := global.GLOAB_CORN.Entry(taskID).Next.Format("2006-01-02 15:04:05")
	g.NextTime = nextTime
	err = global.GLOAB_DB.Updates(&g).Error
	if err != nil {
		zap.L().Error("获取定时任务下一次执行时间有误", zap.Error(err))
		return
	}
	return result, taskID, err
}

//提醒未打卡的同学考勤
func (a *DingAttendGroup) AlertAttent(p *params.ParamAllDepartAttendByRobot) (result map[string][]DingAttendance, taskID cron.EntryID, err error) {
	//判断一下是否需要需要课表小程序的数据
	token, err := (&DingToken{}).GetAccessToken()
	if err != nil || token == "" {
		zap.L().Error("从redis中取出token失败", zap.Error(err))
		return
	}
	g := DingAttendGroup{GroupId: p.GroupId, DingToken: DingToken{Token: token}}
	commutingTime, err := g.GetCommutingTime()
	if err != nil {
		zap.L().Error("根据考勤组获取上下班时间失败", zap.Error(err))
		return
	}
	//获取到上班时间
	OnDutyTimeList := commutingTime["OnDuty"]
	//把时间格式拼装处理一下，拼装成corn定时库spec定时规则能够使用的格式
	min := ""
	hour := ""
	for i := 0; i < len(OnDutyTimeList); i++ {
		s := strings.Split(strings.Split(OnDutyTimeList[i], " ")[1], ":")
		h, _ := strconv.Atoi(s[0])
		m, _ := strconv.Atoi(s[1])
		time, _ := strconv.Atoi(utils.Advance)
		//先转化成分钟
		i2 := h*60 + m
		m = (i2 - time) % 60
		h = (i2 - time) / 60
		minute := strconv.Itoa(m)
		hours := strconv.Itoa(h)
		hour += hours + ","
		min += minute + ","
	}
	hour = hour[:len(hour)-1]
	min = min[:len(min)-1]
	//把时间格式拼装处理一下，拼装成corn定时库spec定时规则能够使用的格式
	spec := "00 " + min + " " + hour + " * * ?"
	//spec = "00 55,25,55 7,14,19 * * ?"
	zap.L().Info(spec)
	task := func() {
		token, err = (&DingToken{}).GetAccessToken()
		g := DingAttendGroup{GroupId: p.GroupId, DingToken: DingToken{Token: token}}
		//a := DingAttendance{DingToken: DingToken{Token: token}}
		//获取一天上下班的时间
		commutingTimes, err := g.GetCommutingTime1()
		if err != nil {
			zap.L().Error("根据考勤组id获取一天上下班失败失败", zap.Error(err))
			return
		}
		//获取上班时间、//获取下班时间
		OnDutyTime := commutingTimes["OnDuty"]
		OffDutyTime := commutingTimes["OffDuty"]
		zap.L().Info(fmt.Sprintf("上班时间：%v", OnDutyTime))
		zap.L().Info(fmt.Sprintf("下班时间：%v", OffDutyTime))
		//获取当前时间，curTime是自己封装的时间类型，有各种格式的时间
		curTime, err := (&localTime.MySelfTime{}).GetCurTime(commutingTimes)
		if err != nil {
			zap.L().Error("获取当前时间失败", zap.Error(err))
			return
		}
		//判断当前时间是否需要运行，我们使用的是cron定时器，corn定时器不支持一些不规则的定时，我们此处做一些判断，跳过一些不合法的时间
		ok := CronHandle(spec, curTime)
		if ok == false {
			zap.L().Info("当前时间cron执行，但是不是我们想要的时间，跳过执行")
			return
		}
		//获取考勤组部门成员，已经筛掉了不参与考勤的个人
		//注意一定要放在task里面，这样当纪检部更新了考勤组之后，每次加载人员都是最新的
		deptAttendanceUser, err := g.GetGroupDeptNumber()
		if err != nil {
			zap.L().Error("获取考勤组部门成员(已经筛掉了不参与考勤的个人)失败", zap.Error(err))
			return
		}
		zap.L().Info(fmt.Sprintf("考勤规则：%v，考勤人员详情：%v", spec, deptAttendanceUser))
		//判断该考勤组是否在校，在校的话，需要判定有课无课情况，如果不在校，则统一按照无课处理
		var isInSchool bool
		err = global.GLOAB_DB.Model(&DingAttendGroup{GroupId: p.GroupId}).Select("is_in_school").Scan(&isInSchool).Error
		if err != nil {
			zap.L().Error("通过考勤组判断是否在学校（加入课表小程序数据失败）", zap.Error(err))
			isInSchool = false
		}
		token, err := (&DingToken{}).GetAccessToken()
		if err != nil {
			zap.L().Error("从redis中取出token失败", zap.Error(err))
			return
		}
		//Len := len(deptAttendanceUser)
		Count := 0
		startWeek, week, CourseNumber, err := DateHandle(curTime)

		for DeptId, _ := range deptAttendanceUser { //
			Count++
			atoi, _ := strconv.Atoi(DeptId)
			DeptDetail, err := (&DingDept{DingToken: DingToken{Token: token}, DeptId: atoi}).GetDeptByIDFromMysql()
			DeptDetail.DingToken.Token = token
			if err != nil {
				zap.L().Error(fmt.Sprintf("通过部门id：%s获取部门详情失败，继续执行下一轮循环", DeptId), zap.Error(err))
				continue
			}
			fmt.Println("当前部门名称", DeptDetail.Name)
			//todo 判断一下此部门是否开启推送考勤
			if DeptDetail.IsRobotAttendance == false || DeptDetail.RobotToken == "" {
				zap.L().Error(fmt.Sprintf("该部门:%s为开启考勤或者机器人robotToken:%s是空，跳过", DeptDetail.Name, DeptDetail.RobotToken))
				continue
			}
			zap.L().Info(fmt.Sprintf("该部门:%s开启考勤,机器人robotToken:%s", DeptDetail.Name, DeptDetail.RobotToken))
			result = make(map[string][]DingAttendance, 0)
			result["Normal"] = make([]DingAttendance, 0)
			result["Late"] = make([]DingAttendance, 0)
			result["Leave"] = make([]DingAttendance, 0)
			result["HasCourse"] = make([]DingAttendance, 0)
			//获取了一个部门所有参与考勤的用户id
			DeptAttendanceUserIdList := GetUserIdListByUserList(deptAttendanceUser[DeptId])
			//根据用户id获取用户打卡情况，同时返回了没有考勤数据的同学
			attendanceList, NotRecordUserIdList, err := DeptDetail.GetAttendanceData(DeptAttendanceUserIdList, curTime, OnDutyTime, OffDutyTime)
			if err != nil {
				zap.L().Error("根据部门用户id列表获取用户打卡情况失败", zap.Error(err))
			}
			fmt.Println("已经打卡的同学", attendanceList)
			//遍历考勤数据,有课的优先级高于考勤
			for _, attendance := range attendanceList {
				flag := false
				//查一下课表，有课且打卡的话，判定为有课
				if isInSchool {
					course, _ := classCourse.GetIsHasCourse(CourseNumber, startWeek, 0, []string{attendance.UserID}, week)
					for _, Byclass := range course {
						if Byclass.Userid == attendance.UserID {
							result["HasCourse"] = append(result["HasCourse"], attendance)
							flag = true
							break
						}
					}
				}
				if flag == false {
					if attendance.TimeResult == "Normal" {
						result["Normal"] = append(result["Normal"], attendance)
					}
				}
			}
			zap.L().Info(fmt.Sprintf("有考勤记录同学已经处理完成，接下来开始处理没有考勤数据的同学"))
			/*
				获取课表小程序有课的同学
				课表小程序有一个接口，可以获取到大家的有课无课情况，其中参数有
				当前周、高级筛选中的部门，我们找到部门中有课的同学，然后跳过即可
			*/
			//处理没有考勤记录的同学，看看其是否有课，map传递的引用类型
			fmt.Println("没有打卡的同学", NotRecordUserIdList)
			if isInSchool {
				//调用课表小程序接口，判断没有考勤数据的人是否请假了
				//需要参数：当前周、周几、第几节课，NotRecordUserIdList
				//此处传递的两个参数 NotRecordUserIdList、result 都是引用类型，NotRecordUserIdList处理之后已经不含有课的成员了
				HasCourseHandle(NotRecordUserIdList, CourseNumber, startWeek, week, result)
			}
			if (week == 1 && curTime.Duration == 3) || (week == 2 && curTime.Duration == 1) || (week == 2 && curTime.Duration == 2) {
				zap.L().Info("freetime跳过")
				//直接所有部门都不再发送了
				return
			}
			//if week == 7 && curTime.Duration == 3 {
			//	zap.L().Info("周日晚上跳过")
			//	//直接所有部门都不再发送了
			//	return
			//}
			//if week == 1 && curTime.Duration == 1 && DeptDetail.DeptId == 440395094 {
			//	zap.L().Info("周一上午三期社招跳过")
			//	//跳过三期校招，继续循环其他部门
			//	continue
			//}
			err, late := LeaveLateHandle(NotRecordUserIdList, token, result, curTime)
			if err != nil {
				zap.L().Error("处理请假和迟到有误", zap.Error(err))
			}

			zap.L().Info("没有考勤数据的同学已经处理完成")
			//将考勤数据发给部门负责人以及管理人员
			p := &ParamChat{
				RobotCode: "dingepndjqy7etanalhi",
				UserIds:   late,
				MsgKey:    "sampleText",
				MsgParam:  "还有五分钟上班，你还没有打卡",
			}
			err = (&DingRobot{}).ChatSendMessage(p)
			if err != nil {
				zap.L().Error("发送提醒信息失败", zap.Error(err))
			}
		}
		return
	}
	//添加一个定时任务
	taskID, err = global.GLOAB_CORN.AddFunc(spec, task)
	if err != nil {
		zap.L().Error("启动机器人查考勤定时任务失败", zap.Error(err))
		return
	}
	nextTime := global.GLOAB_CORN.Entry(taskID).Next.Format("2006-01-02 15:04:05")
	g.NextTime = nextTime
	err = global.GLOAB_DB.Updates(&g).Error
	if err != nil {
		zap.L().Error("获取定时任务下一次执行时间有误", zap.Error(err))
		return
	}
	return result, taskID, err
}

func BitMapHandle(result map[string][]DingAttendance, curTime localTime.MySelfTime, startWeek, weekDay int) (err error) {
	//把有课，打卡的，请假的放入sign切片中
	sign := make([]DingAttendance, 0)
	sign = append(sign, result["Normal"]...)
	sign = append(sign, result["Leave"]...)
	sign = append(sign, result["HasCourse"]...)
	year, _ := strconv.Atoi(curTime.Format[0:4])
	mouth, _ := strconv.Atoi(curTime.Format[5:7])
	//upOrDown用于判断是上半年还是下半年
	upOrDown := 0
	//如果月份小于九，就是上半年
	if mouth < 9 {
		upOrDown = 1
	} else {
		upOrDown = 2
	}
	for i := 0; i < len(sign); i++ {
		//让每一个用户进行签到
		consecutiveSignNum, err := (&DingUser{UserId: sign[i].UserID}).Sign(year, upOrDown, startWeek, weekDay, curTime.Duration)
		if err != nil {
			zap.L().Error("用户打卡后签到存储redis失败", zap.Error(err))
		} else {
			zap.L().Info(fmt.Sprintf("用户打卡后签到存储redis成功，用户%v，连续签到次数：%v", sign[i].UserName, consecutiveSignNum))
		}
	}
	return err
}

func MessageHandle(curTime localTime.MySelfTime, DeptDetail DingDept, result map[string][]DingAttendance) (message string) {
	MANCourseNum := ""
	if curTime.Duration == 1 {
		if curTime.ClassNumber == 1 {
			MANCourseNum = "上午第一节"
		} else if curTime.ClassNumber == 2 {
			MANCourseNum = "上午第二节"
		}

	} else if curTime.Duration == 2 {
		if curTime.ClassNumber == 1 {
			MANCourseNum = "下午第一节"
		} else if curTime.ClassNumber == 2 {
			MANCourseNum = "下午第二节"
		}
	} else if curTime.Duration == 3 {
		MANCourseNum = "晚上"
	}
	message = MANCourseNum + DeptDetail.Name + "考勤结果如下:\n"

	for key, DingAttendance := range result {
		if key == "Normal" {
			message += "正常: "
		} else if key == "Late" {
			message += "迟到: "
		} else if key == "Leave" {
			message += "请假: "
		} else if key == "HasCourse" {
			message += "有课: "
		}
		//下面的循环每次统计一个部门的一种情况
		for _, attendance := range DingAttendance {
			if key == "Leave" {
				//我们把请假的信息给存入到redis中
				//我们使用人名作为key，使用请假次数作为value
			}
			message += attendance.UserName + " "
		}
		message += "\n"
	}
	return message
}

func LeaveLateHandle(NotRecordUserIdList []string, token string, result map[string][]DingAttendance, curTime localTime.MySelfTime) (err error, late []string) {
	var dl DingLeave
	dl.DingToken.Token = token
	limit := 20
	Offset := 0
	hasMore := true
	late = make([]string, len(NotRecordUserIdList))
	//遍历每一个没有考勤记录的同学
	for i := 0; i < len(NotRecordUserIdList); i++ {
		var u DingUser
		u.DingToken.Token = token
		u.UserId = NotRecordUserIdList[i]
		NotAttendanceUser, err := u.GetUserDetailByUserId()
		if err != nil {
			zap.L().Error(fmt.Sprintf("遍历每一个没有考勤记录也没有课的同学的过程中,通过钉钉用户id:%s获取钉钉用户详情失败", NotRecordUserIdList[i]), zap.Error(err))
			continue
		}
		zap.L().Info(fmt.Sprintf("%s没有考勤数据,没有有课信息，接下来开始获取其请假数据", NotAttendanceUser.Name))
		leaveStatusList := make([]DingLeave, 0)
		hasMore = true
		for hasMore {
			zap.L().Info(fmt.Sprintf("姓名：%v提交请假开始时间%v,提交请假结束时间:%v ，把时间戳转化为可以看懂的时间，开始:%s,结束:%s", NotAttendanceUser.Name, curTime.TimeStamp-10*86400000, curTime.TimeStamp, (&localTime.MySelfTime{}).StampToString(curTime.TimeStamp-10*86400000), curTime.Format))
			leaveStatusListSection, HasMore, err := dl.GetLeaveStatus(curTime.TimeStamp-10*86400000, curTime.TimeStamp, Offset, limit, NotRecordUserIdList[i])
			if err != nil {
				zap.L().Error("获取请假状态失败，跳过继续执行下一条数据", zap.Error(err))
				continue
			}
			leaveStatusList = append(leaveStatusList, leaveStatusListSection...)
			hasMore = HasMore
			if hasMore {
				Offset = Offset + 1
			}
		}
		leave := DingLeave{}

		if len(leaveStatusList) > 0 {
			sort.Slice(leaveStatusList, func(i, j int) bool {
				return leaveStatusList[i].EndTime > leaveStatusList[j].StartTime
			})
			leave = leaveStatusList[0]
			zap.L().Info(fmt.Sprintf("%v获取到了请假数据，只保留最后一条请假记录，请假生效时间:%v,请假结束时间:%v", NotAttendanceUser.Name, (&localTime.MySelfTime{}).StampToString(leave.StartTime), (&localTime.MySelfTime{}).StampToString(leave.EndTime)))
		}

		if leave.StartTime != 0 && leave.StartTime < curTime.TimeStamp && leave.EndTime > curTime.TimeStamp-utils.Delay*1000 {
			result["Leave"] = append(result["Leave"], DingAttendance{TimeResult: "Leave", CheckType: "OnDuty", UserID: NotRecordUserIdList[i], UserName: NotAttendanceUser.Name})
			zap.L().Info(fmt.Sprintf("%s在合法时间段请假，被判定为请假", NotAttendanceUser.Name))
		} else {
			zap.L().Info(fmt.Sprintf("%s未在合法时间段请假，被判定为迟到", NotAttendanceUser.Name))
			result["Late"] = append(result["Late"], DingAttendance{TimeResult: "Late", CheckType: "OnDuty", UserID: NotRecordUserIdList[i], UserName: NotAttendanceUser.Name})
			//创建一个还未打卡的切片，并返回
			late = append(late, NotRecordUserIdList[i])
		}
		//Todo 把每个人的请假状态的最后一次记录存储到redis中

	}
	return
}

func HasCourseHandle(NotRecordUserIdList []string, CourseNumber int, startWeek int, weekday int, result map[string][]DingAttendance) {
	if len(NotRecordUserIdList) > 0 {
		ByClass, err := classCourse.GetIsHasCourse(CourseNumber, startWeek, 0, NotRecordUserIdList, weekday)
		if err != nil {
			zap.L().Error("获取当前部门参与考勤的人是否有课失败", zap.Error(err))
		}
		for _, class := range ByClass {
			//找到有课同学的下标，然后在NotRecordUserIdList中把该下标对应的元素移除掉
			for i := 0; i < len(NotRecordUserIdList); i++ {
				if NotRecordUserIdList[i] == class.Userid {
					zap.L().Info(fmt.Sprintf("%v没有打卡记录，查询到其有课，属于正常情况", class.UName))
					result["HasCourse"] = append(result["HasCourse"], DingAttendance{TimeResult: "HasCourse", CheckType: "OnDuty", UserID: NotRecordUserIdList[i], UserName: class.UName})
					NotRecordUserIdList = append(NotRecordUserIdList[:i], NotRecordUserIdList[i+1:]...)
					break
				}
			}
		}
	} else {
		zap.L().Info("该部门全部出勤，不再判断是否有课")
	}
}

// SundayAfternoonExec 此函数周日下午执行
func SundayAfternoonExec(startWeek int) {
	r := &DingRobot{RobotId: "aba857cf3ba132581d1a99f3f5c9c5fe2754ffd57a3e7929b6781367b9325e40"}
	// 此函数报告本周的请假情况
	SundayLeaveExec(startWeek, r)
	// 此函数报告本周的迟到情况
	SundayLateExec(startWeek, r)
}

// SundayLeaveExec 此函数报告本周的请假情况
func SundayLeaveExec(startWeek int, r *DingRobot) {
	leaveResult, err := global.GLOBAL_REDIS.ZRevRangeWithScores(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":", 0, 100).Result()
	if err != nil {
		zap.L().Info("平均请假次数排行信息获取失败")
		return
	}
	message := "各部门平均请假次数排行如下:\n"
	for i, v := range leaveResult {
		// 获取此部门名称
		deptName := strings.Split(v.Member.(string), "部门名称:")[1]
		// 获取此部门的请假总次数
		deptCount := global.GLOBAL_REDIS.Get(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":dept:"+deptName).Val()
		// 获取此部门的平均请假次数
		deptAveCount := v.Score
		message += fmt.Sprintf("%v. %v请假总次数为: %v, 平均请假次数为: %v\n", i+1, deptName, deptCount, deptAveCount)
	}
	pSend := &ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{IsAtAll: false},
			Text: common.Text{
				Content: message,
			},
			Msgtype: "text",
		},
	}
	if err = r.SendMessage(pSend); err != nil {
		zap.L().Info("机器人发送平均请假次数排行信息失败")
		return
	}

}

// SundayLateExec 此函数报告本周的迟到情况
func SundayLateExec(startWeek int, r *DingRobot) {
	lateResult, err := global.GLOBAL_REDIS.ZRevRangeWithScores(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":", 0, 100).Result()
	if err != nil {
		zap.L().Info("平均迟到次数排行信息获取失败")
		return
	}
	message := "各部门平均迟到次数排行如下:\n"
	for i, v := range lateResult {
		// 获取此部门名称
		deptName := strings.Split(v.Member.(string), "部门名称:")[1]
		// 获取此部门的迟到总次数
		deptCount := global.GLOBAL_REDIS.Get(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":dept:"+deptName).Val()
		// 获取此部门的平均请假次数
		deptAveCount := v.Score
		message += fmt.Sprintf("%v. %v迟到总次数为: %v, 平均迟到次数为: %v\n", i+1, deptName, deptCount, deptAveCount)
	}
	// 要发送的信息
	pSend := &ParamCronTask{
		MsgText: &common.MsgText{
			At: common.At{IsAtAll: false},
			Text: common.Text{
				Content: message,
			},
			Msgtype: "text",
		},
	}
	if err = r.SendMessage(pSend); err != nil {
		zap.L().Info("机器人发送平均迟到次数排行信息失败")
		return
	}
}

type Message struct {
	DepartmentName string
	FirstName      string
	Time           int64
}

func DeptFirstShowUpMorning(p *params.ParamGetDeptFirstShowUpMorning) (err error) {
	var task Task
	CronTask := func() {
		fmt.Println("进入到了定时任务")
		timeArr, dur, FirstTime, Mtime, Aftime, EvTime, err := TimeTransFrom()
		//Switch()
		fmt.Println(Mtime, Aftime, EvTime)
		if err != nil {
			fmt.Println(err)
			return
		}
		m := map[string]*Message{}
		//根据groupId我们可以获取到参与考勤的部门
		ag := DingAttendGroup{}
		ag.DingToken.Token = p.Token
		ag.GroupId = p.GroupID
		result, err := ag.GetAttendancesGroupMemberList("413550622937553255")
		if err != nil {
			fmt.Println(err)
			return
		}
		//封装一个map，key为部门id，键是一个struct｛departmentName,firstName,time｝

		FinalMessage := dur + "各部门最早打卡人员：\n"
		for _, Member := range result {
			M := Message{}
			if Member.Type == "1" && Member.AtcFlag == "0" { //部门且参与考勤
				m[Member.MemberID] = nil //先填充map的键，值先设置为空
				deptId, _ := strconv.Atoi(Member.MemberID)
				var d DingDept
				d.DingToken.Token = p.Token
				d.DeptId = deptId
				dept, err := d.GetDeptDetailByDeptId()
				if err != nil {
					fmt.Println(err)
				}
				UserList, _, err := d.GetUserListByDepartmentID(0, 100)
				if err != nil {
					fmt.Println(err)
				}
				UserIDList := make([]string, 0)
				for _, val := range UserList {
					UserIDList = append(UserIDList, val.UserId)
				}
				for index := range timeArr {
					attendanceList := make([]DingAttendance, 0)
					//split := UserIdListSplit(UserIDList, 50)
					for i := 0; i <= len(UserIDList)/50; i++ {
						var split []string
						if len(UserIDList) <= (i+1)*50 {
							split = UserIDList[i*50:]
						} else {
							split = UserIDList[i*50 : (i+1)*50]
						}
						if index < len(timeArr)-1 {
							// Todo 修改成模型方法
							attend := DingAttendance{}
							attend.Token = p.Token
							list, err := attend.GetAttendanceList(split, timeArr[index], timeArr[index+1])
							if err != nil {
								fmt.Println(err)
							}
							attendanceList = append(attendanceList, list...)
						}

					}
					//说明此部门在此时间段已经有数据了，不用再找后面的数据了
					if len(attendanceList) != 0 {
						for _, list := range attendanceList {
							if list.UserCheckTime < FirstTime {

								M.Time = list.UserCheckTime
								// Todo 修改成模型方法
								u := DingUser{
									UserId: list.UserID,
									DingToken: DingToken{
										Token: p.Token,
									},
								}
								user, err := u.GetUserDetailByUserId()
								if err != nil {
									fmt.Println(err)

								}

								M.FirstName = user.Name
								M.DepartmentName = dept.Name
							}
						}
						m[Member.MemberID] = &M
						FinalMessage += MessageToString(&M)
						break
					}
				}
				//一个部门的数据已经出来了
			}
		}
		d := DingRobot{
			RobotId: utils.TestRobotToken,
		}
		pSend := &ParamCronTask{
			MsgText: &common.MsgText{
				At: common.At{IsAtAll: true},
				Text: common.Text{
					Content: FinalMessage,
				},
				Msgtype: "text",
			},
		}
		err = d.SendMessage(pSend)
		if err != nil {
			fmt.Println(err)
			return
		}

	}
	//userID, _ := global.GetCurrentUserID(c)
	//userName, _ := global.GetCurrentUserName(c)

	task = Task{
		TaskName: " 早上推送各部门最早出勤人员",
		//UserID:    userID,
		//UserName:  userName,
		RobotId:   utils.TestRobotToken,
		RobotName: utils.LeZhiAllPeopleRobotName,
		MsgText: &common.MsgText{
			Text: common.Text{
				Content: "根据考勤情况而定",
			},
			At: common.At{
				IsAtAll: true,
			},
		},
	}
	MorningEntryID, err := global.GLOAB_CORN.AddFunc(utils.SpecMorning, CronTask)
	if err != nil {
		return
	}
	task.Spec = utils.SpecMorning
	task.TaskID = strconv.Itoa(int(MorningEntryID))
	//err = mysql.InsertTask(task)
	if err != nil {
		return
	}

	AfternoonEntryID, err := global.GLOAB_CORN.AddFunc(utils.SpecAfternoon, CronTask)

	task.TaskName = "下午推送各部门最早出勤人员"
	task.Spec = utils.SpecAfternoon
	task.TaskID = strconv.Itoa(int(AfternoonEntryID))
	//err = mysql.InsertTask(task)
	if err != nil {
		return
	}

	EveningEntryID, err := global.GLOAB_CORN.AddFunc(utils.SpecEvening, CronTask)

	task.Spec = utils.SpecEvening
	task.TaskName = "晚上推送各部门最早出勤人员"
	task.TaskID = strconv.Itoa(int(EveningEntryID))
	//err = mysql.InsertTask(task)
	if err != nil {
		return
	}
	return
}
func GetUserIdListByUserList(UserList []DingUser) (UserIdList []string) {
	for _, val := range UserList {
		UserIdList = append(UserIdList, val.UserId)
	}
	return UserIdList
}
func TimeTransFrom() (target []string, Duration string, FirstTime int64, MorningTime, AfternoonTime, EveningTime int64, err error) {
	CurDate := time.Now()
	CurTimeString := CurDate.Format("15:04:05")
	formatTime, _ := time.Parse("15:04:05", CurTimeString)
	times := [][]int{{0, 0, 7, 0, 8, 0, 8, 20, 8, 30}, {11, 30, 13, 30, 14, 0}, {18, 0, 19, 0, 19, 30}}
	//当前时间在上午，times[0],下午1，晚上2
	//Todo 判断时间在哪个时间段
	x := 0
	if formatTime.Before(utils.AttendanceMorningTime) {
		x = 0
		Duration = utils.Morning
		FirstTime = time.Date(CurDate.Year(), CurDate.Month(), CurDate.Day(), times[x][len(times[x])-2], times[x][len(times[x])-1], 0, 0, time.Local).Unix() * 1000
	} else if formatTime.Before(utils.AttendanceAfternoonTime) {
		x = 1
		Duration = utils.Afternoon
		FirstTime = time.Date(CurDate.Year(), CurDate.Month(), CurDate.Day(), times[x][len(times[x])-2], times[x][len(times[x])-1], 0, 0, time.Local).Unix() * 1000
	} else if formatTime.Before(utils.AttendanceEveningTime) {
		x = 2
		Duration = utils.Evening
		FirstTime = time.Date(CurDate.Year(), CurDate.Month(), CurDate.Day(), times[x][len(times[x])-2], times[x][len(times[x])-1], 0, 0, time.Local).Unix() * 1000
	}

	for i := 0; i < len(times[x])-1; i += 2 {
		target = append(target, time.Date(CurDate.Year(), CurDate.Month(), CurDate.Day(), times[x][i], times[x][i+1], 0, 0, time.Local).Format("2006-01-02 15:04:05"))
	}
	return
}
func MessageToString(m *Message) string {
	tran := timestampTran("15:04:05", m.Time)
	return fmt.Sprintf("部门：%s,人员：%s,打卡时间：%v\n", m.DepartmentName, m.FirstName, tran)
}
func timestampTran(format string, t int64) (s string) {
	t = t / 1000
	s = time.Unix(t, 0).Format("2006:01:02 15:04:05")
	if format == "15:04:05" {
		return s[len(s)-8 : len(s)]
	} else if format == "2006:01:02 15:04:05" {
		return s
	} else {
		return s
	}
}
