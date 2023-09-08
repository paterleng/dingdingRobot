package params

import "ding/model/common/request"

type ParamGetAccessToken struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}
type ParamGetDepartmentList struct {
	Token string `json:"token"`
}
type ParamGetAttendanceGroups struct {
	Token  string `json:"token"`
	Offset int    `json:"offset"`
	Size   int    `json:"size"`
}
type ParamGetDepartmentListByID struct {
	ID    int    `json:"id"` // 部门id，如果最初是1，则代表是根部门
	Token string `json:"token"`
}
type ParamGetDepartmentListByID2 struct {
	ID int `json:"id" form:"id"` // 部门id，如果最初是1，则代表是根部门
}
type ParamGetDeptListFromMysql struct {
	request.PageInfo
}
