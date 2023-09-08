package dingding

import (
	"context"
	"crypto/tls"
	"ding/global"
	"ding/initialize/redis"
	"ding/model/common"
	"ding/model/common/localTime"
	"ding/model/params"
	"ding/model/params/ding"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DingDept struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	UserList  []DingUser `gorm:"many2many:user_dept"`
	DeptId    int        `gorm:"primaryKey" json:"dept_id"`
	Deleted   gorm.DeletedAt
	Name      string `json:"name"`
	ParentId  int    `json:"parent_id"`
	DingToken
	IsSendFirstPerson int        `json:"is_send_first_person"` // 0为不推送，1为推送
	RobotToken        string     `json:"robot_token"`
	IsRobotAttendance bool       `json:"is_robot_attendance"` //是否
	IsJianShuOrBlog   int        `json:"is_jianshu_or_blog" gorm:"column:is_jianshu_or_blog"`
	IsLeetCode        int        `json:"is_leet_code"`
	ResponsibleUsers  []DingUser `gorm:"-"`
}
type UserDept struct {
	DingUserUserID string
	DingDeptDeptID string
	IsResponsible  bool
	Deleted        gorm.DeletedAt
}

//自定义表名建表
func (UserDept) user_dept() string {
	return "user_dept"
}

//获取用户的考勤信息
func (d *DingDept) GetAttendanceData(userids []string, curTime localTime.MySelfTime, OnDutyTime []string, OffDutyTime []string) (attendanceList []DingAttendance, NotRecordUserIdList []string, err error) {
	attendanceList = make([]DingAttendance, 0)
	a := DingAttendance{DingToken: DingToken{Token: d.Token}}
	if userids != nil || len(userids) != 0 {
		for i := 0; i <= len(userids)/50; i++ {
			var split []string
			if len(userids) <= (i+1)*50 {
				split = userids[i*50:]
			} else {
				split = userids[i*50 : (i+1)*50]
			}
			var list []DingAttendance
			zap.L().Info(fmt.Sprintf("接下来开始获取考勤数据，当前时间为：%v %s", curTime.Duration, curTime.Time))
			fmt.Println("curTime.Duration = ", curTime.Duration)
			if len(OnDutyTime) == 3 {
				if curTime.Duration == 1 {
					zap.L().Info(fmt.Sprintf("获取上午考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]))
					list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 2 {
					zap.L().Info(fmt.Sprintf("获取下午考勤数据,userIds:%v,开始时间%s,结束时间：%s ", split, OffDutyTime[0], OnDutyTime[1]))
					list, err = a.GetAttendanceList(split, OffDutyTime[0], OnDutyTime[1])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[0], OnDutyTime[1]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 3 {
					zap.L().Info(fmt.Sprintf("获取晚上考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, OffDutyTime[1], OnDutyTime[2]))
					list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[2])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[1], OnDutyTime[2]), zap.Error(err))
						continue
					}
				}

				if len(list) == 0 {
					zap.L().Error("第一次获取考勤数据长度为0，再获取一次")
					if curTime.Duration == 1 {
						list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0])
						if err != nil {
							zap.L().Error(fmt.Sprintf("第二次获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]), zap.Error(err))
							continue
						}
					} else if curTime.Duration == 2 {
						list, err = a.GetAttendanceList(split, OffDutyTime[0], OnDutyTime[1])
						if err != nil {
							zap.L().Error(fmt.Sprintf("第二次获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[0], OnDutyTime[1]), zap.Error(err))
							continue
						}
					} else if curTime.Duration == 3 {
						list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[2])
						if err != nil {
							zap.L().Error(fmt.Sprintf("第二次获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[1], OnDutyTime[2]), zap.Error(err))
							continue
						}
					}
					if len(list) == 0 {
						zap.L().Error("第二次获取考勤数据仍然为空")
						continue
					}
				}
				zap.L().Info(fmt.Sprintf("部门：%s,成功获取%v,考勤数据，具体数据:%v", d.Name, curTime.Duration, list))
			} else if len(OnDutyTime) == 5 {
				//如果是第二节课考勤
				if curTime.Duration == 1 {
					if curTime.ClassNumber == 1 {
						zap.L().Info(fmt.Sprintf("获取上午第一节课考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]))
						list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[0])
					} else if curTime.ClassNumber == 2 {
						//如果连续的两节课都是没课，第一节课打过卡的不用再打了
						zap.L().Info(fmt.Sprintf("获取上午第二节课考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, curTime.Format[:10]+" 00:00:00", OnDutyTime[1]))
						list, err = a.GetAttendanceList(split, curTime.Format[:10]+" 00:00:00", OnDutyTime[1])
					}
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, curTime.Format[:10]+" 00:00:00", OnDutyTime[0]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 2 {
					if curTime.ClassNumber == 1 {
						zap.L().Info(fmt.Sprintf("获取下午考勤数据,userIds:%v,开始时间%s,结束时间：%s ", split, OffDutyTime[1], OnDutyTime[2])) //第二个下班时间到第三个下班时间
						list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[2])
					} else if curTime.ClassNumber == 2 {
						zap.L().Info(fmt.Sprintf("获取下午考勤数据,userIds:%v,开始时间%s,结束时间：%s ", split, OffDutyTime[1], OnDutyTime[3])) //第二个下班时间到第三个下班时间
						//如果连续的两节课都是没课，第一节课打过卡的不用再打了
						list, err = a.GetAttendanceList(split, OffDutyTime[1], OnDutyTime[3])
					}

					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[0], OnDutyTime[1]), zap.Error(err))
						continue
					}
				} else if curTime.Duration == 3 {
					zap.L().Info(fmt.Sprintf("获取晚上考勤数据,userIds:%v,开始时间%s,结束时间：%s", split, OffDutyTime[3], OnDutyTime[4]))
					list, err = a.GetAttendanceList(split, OffDutyTime[3], OnDutyTime[4])
					if err != nil {
						zap.L().Error(fmt.Sprintf("获取考勤数据失败,失败部门:%s，获取考勤时间范围:%s-%s", d.Name, OffDutyTime[3], OnDutyTime[4]), zap.Error(err))
						continue
					}
				}
			}
			attendanceList = append(attendanceList, list...)
			//处理该部门获取到的考勤记录，只保留上班打卡的记录
		}
		for i := 0; i < len(attendanceList); i++ {
			if attendanceList[i].CheckType == "OffDuty" {
				attendanceList = append(attendanceList[:i], attendanceList[i+1:]...)
				//数组变了，把下标往后移动回去
				i--
			}
		}
		zap.L().Info("只保留上班打卡数据成功")
	}
	HasAttendanceDateUser := make(map[string]int64, 0)
	for i := 0; i < len(attendanceList); i++ {
		u := DingUser{
			DingToken: DingToken{
				Token: d.Token,
			},
			UserId: attendanceList[i].UserID,
		}
		user, err := u.GetUserDetailByUserId()
		if err != nil {
			zap.L().Error(fmt.Sprintf("考勤数据中的成员id:%s 转化为详细信息失败", attendanceList[i].UserID), zap.Error(err))
			continue
		}
		attendanceList[i].UserName = user.Name //完善考勤记录
		HasAttendanceDateUser[attendanceList[i].UserID] = attendanceList[i].UserCheckTime
	}
	zap.L().Info(fmt.Sprintf("打卡机数据获取完毕，完成数据如下：%v", attendanceList))
	NotRecordUserIdList = make([]string, 0)
	for _, UserId := range userids {
		//找到没有考勤记录的人
		_, ok := HasAttendanceDateUser[UserId]
		if !ok {
			NotRecordUserIdList = append(NotRecordUserIdList, UserId)
		}
	}
	return
}
func (d *DingDept) SendFrequencyLeave(startWeek int) error {
	//从redis中取数据，封装，调用钉钉接口，发送即可
	key := redis.KeyDeptAveLeave + strconv.Itoa(startWeek) + ":dept:" + d.Name + ":detail:"
	//global.GLOBAL_REDIS.ZRevRange().Result()
	//从小到达进行排序
	//results, err := global.GLOBAL_REDIS.ZRevRange(context.Background(), key, 0, -1).Result()
	results, err := global.GLOBAL_REDIS.ZRangeWithScores(context.Background(), key, 0, -1).Result()
	if err != nil {

	}
	msg := d.Name + "请假情况如下：\n"
	for i := 0; i < len(results); i++ {
		name := results[i].Member.(string)
		time := int(results[i].Score)
		msg += name + "请假次数：" + strconv.Itoa(time) + "\n"
	}

	p := &ParamCronTask{
		MsgText: &common.MsgText{
			Msgtype: "text",
			Text:    common.Text{Content: msg},
		},
		RepeatTime: "立即发送",
	}
	(&DingRobot{RobotId: "aba857cf3ba132581d1a99f3f5c9c5fe2754ffd57a3e7929b6781367b9325e40"}).CronSend(nil, p)
	return nil
}
func (d *DingDept) CountFrequencyLeave(startWeek int, result map[string][]DingAttendance) (err error) {
	//我们取到所有请假的同学，然后进行登记
	for i := 0; i < len(result["Leave"]); i++ {
		//对部门中的每一位同学进行统计
		//NX可以不存在时创建，存在时更新，ZIncrBy的话，可以以固定数值加分，如果是Z
		err = global.GLOBAL_REDIS.ZIncrBy(context.Background(), redis.KeyDeptAveLeave+strconv.Itoa(startWeek)+":dept:"+d.Name+":detail:", 1, result["Leave"][i].UserName).Err()
		if err != nil {
			fmt.Println(err)
		}
	}
	//此处应该被复用一下
	return
}
func (d *DingDept) SendFrequencyLate(startWeek int) error {
	//从redis中取数据，封装，调用钉钉接口，发送即可
	key := redis.KeyDeptAveLate + strconv.Itoa(startWeek) + ":dept:" + d.Name + ":detail:"
	//global.GLOBAL_REDIS.ZRevRange().Result()
	//从小到达进行排序
	//results, err := global.GLOBAL_REDIS.ZRevRange(context.Background(), key, 0, -1).Result()
	results, err := global.GLOBAL_REDIS.ZRangeWithScores(context.Background(), key, 0, -1).Result()
	if err != nil {

	}
	msg := d.Name + "迟到次数如下：\n"
	for i := 0; i < len(results); i++ {
		name := results[i].Member.(string)
		time := int(results[i].Score)
		msg += name + "迟到次数：" + strconv.Itoa(time) + "\n"
	}
	fmt.Println("发送迟到频率了")
	p := &ParamCronTask{
		MsgText: &common.MsgText{
			Msgtype: "text",
			Text:    common.Text{Content: msg},
		},
		RepeatTime: "立即发送",
	}
	(&DingRobot{RobotId: "aba857cf3ba132581d1a99f3f5c9c5fe2754ffd57a3e7929b6781367b9325e40"}).CronSend(nil, p)
	return nil
}
func (d *DingDept) CountFrequencyLate(startWeek int, result map[string][]DingAttendance) (err error) {
	//我们取到所有请假的同学，然后进行登记
	for i := 0; i < len(result["Late"]); i++ {
		//对部门中的每一位同学进行统计
		//NX可以不存在时创建，存在时更新，ZIncrBy的话，可以以固定数值加分，如果是Z
		err = global.GLOBAL_REDIS.ZIncrBy(context.Background(), redis.KeyDeptAveLate+strconv.Itoa(startWeek)+":dept:"+d.Name+":detail:", 1, result["Late"][i].UserName).Err()
		if err != nil {
			fmt.Println(err)
		}
	}
	//此处应该被复用一下
	return
}

type JinAndBlogClassify struct {
	DeptId int          `gorm:"primaryKey" json:"dept_id"`
	Name   string       `json:"name"`
	Data   []JinAndBlog `json:"data" gorm:"many2many:user_dept"`
}

func (d *DingDept) GetAllJinAndBlog() (result []JinAndBlogClassify, err error) {

	var DeptList []DingDept
	err = global.GLOAB_DB.Model(&DingDept{}).Where("is_jianshu_or_blog = 1").Select("dept_id", "name").Preload("UserList").Find(&DeptList).Error
	result = make([]JinAndBlogClassify, len(DeptList))
	for i := 0; i < len(DeptList); i++ {
		result[i].Name = DeptList[i].Name
		result[i].DeptId = DeptList[i].DeptId
		result[i].Data = make([]JinAndBlog, len(DeptList[i].UserList))
		for j := 0; j < len(DeptList[i].UserList); j++ {
			result[i].Data[j].Name = DeptList[i].UserList[j].Name
			result[i].Data[j].UserId = DeptList[i].UserList[j].UserId
			result[i].Data[j].JianShuArticleURL = DeptList[i].UserList[j].JianShuArticleURL
			result[i].Data[j].BlogArticleURL = DeptList[i].UserList[j].BlogArticleURL
		}
	}
	return

}

//通过部门id获取部门用户详情 https://open.dingtalk.com/document/isvapp/queries-the-complete-information-of-a-department-user
func (d *DingDept) GetUserListByDepartmentID(cursor, size int) (userList []DingUser, hasMore bool, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/list?access_token=" + d.DingToken.Token
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
		DeptID: d.DeptId,
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
		return nil, false, errors.New(r.Errmsg)
	}

	// 此处举行具体的逻辑判断，然后返回即可
	return r.Result.List, r.Result.HasMore, nil
}

//两个数组取差集
func DiffArray(a []DingDept, b []DingDept) []DingDept {
	var diffArray []DingDept
	temp := map[int]struct{}{}

	for _, val := range b {
		if _, ok := temp[val.DeptId]; !ok {
			temp[val.DeptId] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val.DeptId]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}
func DiffSilceDept(a []DingDept, b []DingDept) []DingDept {
	var diffArray []DingDept
	temp := map[int]struct{}{}

	for _, val := range b {
		if _, ok := temp[val.DeptId]; !ok {
			temp[val.DeptId] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val.DeptId]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}
func DiffSilceUser(a []DingUser, b []DingUser) []DingUser {
	var diffArray []DingUser
	temp := map[string]struct{}{}

	for _, val := range b {
		if _, ok := temp[val.UserId]; !ok {
			temp[val.UserId] = struct{}{}
		}
	}

	for _, val := range a {
		if _, ok := temp[val.UserId]; !ok {
			diffArray = append(diffArray, val)
		}
	}

	return diffArray
}

//递归查询部门并存储到数据库中
func (d *DingDept) ImportDeptData() (DepartmentList []DingDept, err error) {
	var oldDept []DingDept
	err = global.GLOAB_DB.Find(&oldDept).Error
	if err != nil {

	}
	var dfs func(string, int) (err error)
	dfs = func(token string, id int) (err error) {
		var client *http.Client
		var request *http.Request
		var resp *http.Response
		var body []byte
		URL := "https://oapi.dingtalk.com/topapi/v2/department/listsub?access_token=" + token
		client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}, Timeout: time.Duration(time.Second * 5)}
		//此处是post请求的请求题，我们先初始化一个对象
		b := struct {
			DeptID int `json:"dept_id"`
		}{
			DeptID: id,
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
			Result []DingDept `json:"result"`
		}{}
		//把请求到的结构反序列化到专门接受返回值的对象上面
		err = json.Unmarshal(body, &r)
		if err != nil {
			return
		}
		if r.Errcode != 0 {
			return errors.New(r.Errmsg)
		}
		// 此处举行具体的逻辑判断，然后返回即可
		subDepartments := r.Result
		DepartmentList = append(DepartmentList, subDepartments...)
		if len(subDepartments) > 0 {
			for i := range subDepartments {
				departmentList := make([]DingDept, 0)
				dfs(token, subDepartments[i].DeptId)
				if err != nil {
					return
				}
				DepartmentList = append(DepartmentList, departmentList...)
			}
		}
		return
	}
	err = dfs(d.DingToken.Token, 1)
	if err != nil {
		return
	}
	//取差集查看一下那些部门已经不在来了，进行软删除
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "dept_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "parent_id"}),
	}).Create(&DepartmentList).Error
	//找到不存在的部门进行软删除,同时删除其关系
	Deleted := DiffSilceDept(oldDept, DepartmentList)
	err = global.GLOAB_DB.Select(clause.Associations).Delete(&Deleted).Error
	//根据部门id存储一下部门用户
	for i := 0; i < len(DepartmentList); i++ {
		UserList := make([]DingUser, 0)
		//调用钉钉接口，获取部门中的成员，然后存储进来
		hasMore := true
		for hasMore {
			tempUserList := make([]DingUser, 0)
			d.DeptId = DepartmentList[i].DeptId
			tempUserList, hasMore, err = d.GetUserListByDepartmentID(0, 100)
			if err != nil {
				zap.L().Error("获取部门用户详情失败", zap.Error(err))
			}
			UserList = append(UserList, tempUserList...)
			fmt.Println(i)
			fmt.Println(hasMore)
		}
		//查到用户后，同步到数据库里面，把不在组织架构里面直接删除掉
		//先查一下老的
		oldUserList := make([]DingUser, 0)
		global.GLOAB_DB.Model(&DingDept{DeptId: DepartmentList[i].DeptId}).Association("UserList").Find(&oldUserList)
		//取差集找到需要删除的名单
		userDeleted := DiffSilceUser(oldUserList, UserList)
		err = global.GLOAB_DB.Select(clause.Associations).Delete(&userDeleted).Error
		err = global.GLOAB_DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "title"}),
		}).Create(&UserList).Error
		//更新用户部门关系，更新的原理是：先把之前该部门的关系全部删除，然后重新添加
		err = global.GLOAB_DB.Model(&DepartmentList[i]).Association("UserList").Replace(UserList)
	}
	return
}

//根据id获取子部门列表详情
func (d *DingDept) GetDepartmentListByID() (subDepartments []DingDept, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/department/listsub?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
	}{
		DeptID: d.DeptId,
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
		Result []DingDept `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New("token有误，尝试输入新token")
	}
	// 此处举行具体的逻辑判断，然后返回即可
	subDepartments = r.Result
	return subDepartments, nil
}

//根据id获取子部门列表详情（从数据库查）
func (d *DingDept) GetDepartmentListByID2() (subDepartments []DingDept, err error) {
	err = global.GLOAB_DB.Where("parent_id = ?", d.DeptId).Find(&subDepartments).Error
	return
}
func (d *DingDept) GetDeptByIDFromMysql() (dept DingDept, err error) {
	err = global.GLOAB_DB.First(&dept, d.DeptId).Error
	return
}
func (d *DingDept) GetDeptByListFromMysql(p *params.ParamGetDeptListFromMysql) (deptList []DingDept, total int64, err error) {
	limit := p.PageSize
	offset := p.PageSize * (p.Page - 1)
	err = global.GLOAB_DB.Limit(limit).Offset(offset).Find(&deptList).Error
	if err != nil {
		zap.L().Error("查询部门列表失败", zap.Error(err))
	}
	err = global.GLOAB_DB.Model(&DingDept{}).Count(&total).Error
	if err != nil {
		zap.L().Error("查询部门列表失败", zap.Error(err))
	}
	return
}

//查看部门推送情况开启推送情况
func (d *DingDept) SendFirstPerson(cursor, size int) {
	var depts []DingDept
	global.GLOAB_DB.Select("Name").Find(&depts)
}

//通过部门id获取部门详细信息（取钉钉接口）  https://open.dingtalk.com/document/isvapp-server/industry-address-book-api-for-obtaining-department-information
func (d *DingDept) GetDeptDetailByDeptId() (dept DingDept, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/department/get?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		DeptID int `json:"dept_id"`
	}{
		DeptID: d.DeptId,
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
		Dept DingDept `json:"result"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return r.Dept, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	return r.Dept, nil
}

//更新部门信息
func (d *DingDept) UpdateDept(p *ding.ParamUpdateDeptToCron) (err error) {
	dept := &DingDept{DeptId: p.DeptID, IsSendFirstPerson: p.IsSendFirstPerson, IsRobotAttendance: p.IsRobotAttendance, RobotToken: p.RobotToken, IsJianShuOrBlog: p.IsJianshuOrBlog}
	err = global.GLOAB_DB.Preload("ResponsibleUsers").Updates(dept).Error
	return err
}
func (d *DingAttendGroup) UpdateSchool(s *ding.ParameIsInSchool) (err error) {
	err = global.GLOAB_DB.Model(&DingAttendGroup{}).Where("group_id", s.GroupId).Update("is_in_school", s.IsInSchool).Error
	return
}
