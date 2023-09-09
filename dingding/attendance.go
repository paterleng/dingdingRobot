package dingding

import (
	"crypto/tls"
	"ding/model/dingding"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ResponseGetAttendancesGroups struct {
	Errcode int                        `json:"errcode"`
	Result  ResultGetAttendancesGroups `json:"result"`
}
type ResultGetAttendancesGroups struct {
	Groups []dingding.DingAttendGroup `json:"groups"`
}

type BodyGetAttendancesGroups struct {
	Offset int `json:"offset"`
	Size   int `json:"size"`
}

//官方文档：https://open-dev.dingtalk.com/apiExplorer?spm=ding_open_doc.document.0.0.2f3645a1HPhgVp#/?devType=org&api=dingtalk.oapi.attendance.getsimplegroups

type BodyGetAttendancesGroupMemberList struct {
	OpUserID string `json:"op_user_id"`
	GroupID  int    `json:"group_id"`
}

type ResponseGetAttendancesGroupMemberList struct {
	Errcode                             int                                 `json:"errcode"`
	ResultGetAttendancesGroupMemberList ResultGetAttendancesGroupMemberList `json:"result"`
}
type ResultGetAttendancesGroupMemberList struct {
	ResultGetAttendancesGroupMemberListResults []ResultGetAttendancesGroupMemberListResult `json:"result"`
}
type ResultGetAttendancesGroupMemberListResult struct {
	AtcFlag  string `json:"atc_flag"`
	Type     string `json:"type"`
	MemberID string `json:"member_id"`
}

//获取考勤组人员（部门id和成员id）https://open.dingtalk.com/document/isvapp-server/batch-query-of-attendance-group-members
func GetAttendancesGroupMemberList(token, OpUserID string, GroupID int) (R []ResultGetAttendancesGroupMemberListResult, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/attendance/group/member/list?access_token=" + token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := BodyGetAttendancesGroupMemberList{
		OpUserID: OpUserID,
		GroupID:  GroupID,
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
	r := ResponseGetAttendancesGroupMemberList{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	// 此处举行具体的逻辑判断，然后返回即可
	R = r.ResultGetAttendancesGroupMemberList.ResultGetAttendancesGroupMemberListResults
	return R, nil
}

type BodyGetAttendanceList struct {
	CheckDateFrom string   `json:"checkDateFrom"`
	CheckDateTo   string   `json:"checkDateTo"`
	UserIds       []string `json:"userIds"`
}

type Recordresult struct {
	UserCheckTime int64  `json:"userCheckTime"` //时间戳实际打卡时间。
	TimeResult    string `json:"timeResult"`    //打卡结果Normal：正常 Early：早退 Late：迟到 SeriousLate：严重迟到 Absenteeism：旷工迟到 NotSigned：未打卡
	CheckType     string `json:"checkType"`     //Ondusty 上班，OffDuty下班
	UserID        string `json:"userId"`
}

type ResponseGetAttendanceList struct {
	dingding.DingResponseCommon
	Recordresults []dingding.DingAttendance `json:"recordresult"`
}

//获取考勤结果（可以根据userid批量查询） https://open.dingtalk.com/document/orgapp-server/open-attendance-clock-in-data
func GetAttendanceList(token string, userIds []string, CheckDateFrom string, CheckDateTo string) (Recordresult []dingding.DingAttendance, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/attendance/listRecord?access_token=" + token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := BodyGetAttendanceList{
		CheckDateFrom: CheckDateFrom,
		CheckDateTo:   CheckDateTo,
		UserIds:       userIds,
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
	r := ResponseGetAttendanceList{}

	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return nil, errors.New(r.Errmsg)
	}

	// 此处举行具体的逻辑判断，然后返回即可

	return r.Recordresults, nil
}

//获取考勤组中的部门成员
func GetGroupDeptNumber(token string, groupId int) (DeptUsers map[string][]dingding.DingUser) {
	DeptUsers = make(map[string][]dingding.DingUser)
	result, err := GetAttendancesGroupMemberList(token, "413550622937553255", groupId)
	NotAttendanceUserIdListMap := make(map[string]struct{})
	for _, Member := range result {
		if Member.Type == "0" && Member.AtcFlag == "1" { //单个人且不参与考勤
			NotAttendanceUserIdListMap[Member.MemberID] = struct{}{}
		}
	}
	for _, Member := range result {
		DeptAttendanceUserList := make([]dingding.DingUser, 0)
		if Member.Type == "1" && Member.AtcFlag == "0" { //部门且参与考勤
			deptId, _ := strconv.Atoi(Member.MemberID)
			//dept, err := GetDeptDetailByDeptId(token, deptId)
			if err != nil {
				return
			}
			var d dingding.DingDept
			d.DingToken.Token = token
			d.DeptId = deptId
			DeptAllUserList, _, err := d.GetUserListByDepartmentID(0, 100)
			if err != nil {
				return
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
	return DeptUsers
}
