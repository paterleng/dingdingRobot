package system

import (
	"ding/model/common/response"
	request "ding/model/params/system"
	"ding/model/system"

	"ding/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateSysDataDictionary(c *gin.Context) {
	var dictionary system.SysDataDictionary
	err := c.ShouldBindJSON(&dictionary)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = utils.Verify(dictionary, utils.DictionaryVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var d system.SysDataDictionary
	err = d.CreateSysDataDictionary(dictionary)
	if err != nil {
		zap.L().Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}
func DeleteSysDataDictionary(c *gin.Context) {
	var dictionary system.SysDataDictionary
	err := c.ShouldBindJSON(&dictionary)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = utils.Verify(dictionary.Model, utils.IdVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var d system.SysDataDictionary
	err = d.DeleteSysDataDictionary(dictionary)
	if err != nil {
		zap.L().Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
func UpdateSysDataDictionary(c *gin.Context) {
	var dictionary system.SysDataDictionary
	err := c.ShouldBindJSON(&dictionary)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = utils.Verify(dictionary.Model, utils.IdVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(dictionary, utils.DictionaryVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var d system.SysDataDictionary
	err = d.UpdateSysDataDictionary(&dictionary)
	if err != nil {
		zap.L().Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}
func FindSysDataDictionary(c *gin.Context) {
	var dictionary system.SysDataDictionary
	err := c.ShouldBindQuery(&dictionary)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 由于id和type是二必选一，因此这里做两次参数校验,如果同时为空则返回异常响应
	idErr := utils.Verify(dictionary.Model, utils.IdVerify)
	typeErr := utils.Verify(dictionary, utils.DictionaryTypeVerify)
	if idErr != nil && typeErr != nil {
		response.FailWithMessage(idErr.Error(), c)
		return
	}
	var d system.SysDataDictionary
	sysDataDictionary, err := d.GetSysDataDictionary(dictionary.Type, dictionary.ID, dictionary.Status)
	if err != nil {
		zap.L().Error("字典未创建或未开启!", zap.Error(err))
		response.FailWithMessage("字典未创建或未开启", c)
		return
	}
	response.OkWithDetailed(gin.H{"resysDataDictionary": sysDataDictionary}, "查询成功", c)
}
func GetSysDataDictionaryList(c *gin.Context) {
	var pageInfo request.SysDataDictionarySearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(pageInfo.PageInfo, utils.PageInfoVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var d system.SysDataDictionary
	list, total, err := d.GetSysDataDictionaryInfoList(pageInfo)
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
