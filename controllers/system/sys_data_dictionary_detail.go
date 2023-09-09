package system

import (
	"ding/model/common/response"
	request "ding/model/params/system"
	"ding/model/system"
	"ding/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateSysDataDictionaryDetail(c *gin.Context) {
	var detail system.SysDataDictionaryDetail
	err := c.ShouldBindJSON(&detail)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var dt system.SysDataDictionaryDetail
	err = dt.CreateSysDataDictionaryDetail(detail)
	if err != nil {
		zap.L().Error("分类字典详情创建失败!", zap.Error(err))
		response.FailWithMessage("分类字典详情创建失败", c)
		return
	}
	response.OkWithMessage("分类字典详情创建成功", c)
}

// DeleteSysDataDictionaryDetail
// @Tags      SysDataDictionaryDetail
// @Summary   删除SysDataDictionaryDetail
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDataDictionaryDetail     true  "SysDataDictionaryDetail模型"
// @Success   200   {object}  response.Response{msg=string}  "删除SysDataDictionaryDetail"
// @Router    /sysDataDictionaryDetail/deleteSysDataDictionaryDetail [delete]
func DeleteSysDataDictionaryDetail(c *gin.Context) {
	var detail system.SysDataDictionaryDetail
	err := c.ShouldBindJSON(&detail)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var dt system.SysDataDictionaryDetail
	err = dt.DeleteSysDataDictionaryDetail(detail)
	if err != nil {
		zap.L().Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateSysDataDictionaryDetail
// @Tags      SysDataDictionaryDetail
// @Summary   更新SysDataDictionaryDetail
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDataDictionaryDetail     true  "更新SysDataDictionaryDetail"
// @Success   200   {object}  response.Response{msg=string}  "更新SysDataDictionaryDetail"
// @Router    /sysDataDictionaryDetail/updateSysDataDictionaryDetail [put]
func UpdateSysDataDictionaryDetail(c *gin.Context) {
	var detail system.SysDataDictionaryDetail
	err := c.ShouldBindJSON(&detail)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var dt system.SysDataDictionaryDetail
	err = dt.UpdateSysDataDictionaryDetail(&detail)
	if err != nil {
		zap.L().Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSysDataDictionaryDetail
// @Tags      SysDataDictionaryDetail
// @Summary   用id查询SysDataDictionaryDetail
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     system.SysDataDictionaryDetail                                 true  "用id查询SysDataDictionaryDetail"
// @Success   200   {object}  response.Response{data=map[string]interface{},msg=string}  "用id查询SysDataDictionaryDetail"
// @Router    /sysDataDictionaryDetail/findSysDataDictionaryDetail [get]
func FindSysDataDictionaryDetail(c *gin.Context) {
	var detail system.SysDataDictionaryDetail
	err := c.ShouldBindQuery(&detail)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(detail, utils.IdVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var dt system.SysDataDictionaryDetail
	reSysDataDictionaryDetail, err := dt.GetSysDataDictionaryDetail(int(detail.ID))
	if err != nil {
		zap.L().Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"reSysDataDictionaryDetail": reSysDataDictionaryDetail}, "查询成功", c)
}

// GetSysDataDictionaryDetailList
// @Tags      SysDataDictionaryDetail
// @Summary   分页获取SysDataDictionaryDetail列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.SysDataDictionaryDetailSearch                       true  "页码, 每页大小, 搜索条件"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "分页获取SysDataDictionaryDetail列表,返回包括列表,总数,页码,每页数量"
// @Router    /sysDataDictionaryDetail/getSysDataDictionaryDetailList [get]
func GetSysDataDictionaryDetailList(c *gin.Context) {
	var pageInfo request.SysDataDictionaryDetailSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var dt system.SysDataDictionaryDetail
	list, total, err := dt.GetSysDataDictionaryDetailInfoList(pageInfo)
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
