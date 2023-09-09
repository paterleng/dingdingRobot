package ding

import (
	"ding/model/common/request"
	dingding2 "ding/model/dingding"
	"ding/model/params/ding"
	"ding/response"
	"ding/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 使用官方接口导入考勤组数据到数据库中
func ImportAttendanceGroupData(c *gin.Context) {
	var AG dingding2.DingAttendGroup
	t := dingding2.DingToken{}
	token, err := t.GetAccessToken()
	AG.DingToken.Token = token
	_, err = AG.GetAttendancesGroups(0, 10)
	if err != nil {
		response.FailWithMessage("入到考勤组数据失败", c)
		return
	}
	response.OkWithMessage("导入考勤组数据成功", c)
}
func UpdateAttendanceGroup(c *gin.Context) {
	var p ding.ParamUpdateUpdateAttendanceGroup
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("参数错误", zap.Error(err))
		response.FailWithMessage("参数有误", c)
	}
	if p.GroupId == 0 {
		response.FailWithMessage("考勤名称或者id不能为空", c)
		return
	}
	t := dingding2.DingToken{}
	token, err := t.GetAccessToken()
	if err != nil {
		response.FailWithMessage("钉钉token获取失败！", c)
		return
	}
	var d dingding2.DingAttendGroup
	d.DingToken.Token = token
	d.GroupId = p.GroupId
	err = d.UpdateAttendGroup(&p)
	if err != nil {
		response.FailWithMessage("更新考勤组信息失败！", c)
		return
	}
	response.OkWithMessage("更新考勤组信息成功！", c)
}
func GetAttendanceGroupListFromMysql(c *gin.Context) {
	var pageInfo request.PageInfo
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(pageInfo, utils.PageInfoVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t := dingding2.DingToken{}
	token, err := t.GetAccessToken()
	if err != nil {
		response.FailWithMessage("钉钉token获取失败！", c)
		return
	}
	var d dingding2.DingAttendGroup
	d.DingToken.Token = token

	AttendanceGroupList, err := d.GetAttendanceGroupListFromMysql(&pageInfo)
	if err != nil {
		response.FailWithMessage("获取考勤组数据成功！", c)
		return
	}
	response.OkWithDetailed(AttendanceGroupList, "获取考勤组数据成功！", c)
}
