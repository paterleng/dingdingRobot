package system

import (
	system2 "ding/controllers/system"

	"ding/model/system"
	"github.com/gin-gonic/gin"
)

func SetupSystem(System *gin.RouterGroup) {
	System.GET("/", system2.WelcomeHandler)
	sysDataDictionary := System.Group("sysDataDictionary")
	{
		sysDataDictionary.POST("createSysDataDictionary", system2.CreateSysDataDictionary)   // 新建SysDataDictionary
		sysDataDictionary.DELETE("deleteSysDataDictionary", system2.DeleteSysDataDictionary) // 新建SysDataDictionary
		sysDataDictionary.PUT("updateSysDataDictionary", system2.UpdateSysDataDictionary)    // 新建SysDataDictionary
		sysDataDictionary.GET("findSysDataDictionary", system2.FindSysDataDictionary)        // 新建SysDataDictionary
		sysDataDictionary.GET("getSysDataDictionaryList", system2.GetSysDataDictionaryList)  // 新建SysDataDictionary
	}
	sysDataDictionaryDetail := System.Group("sysDataDictionaryDetail")
	{
		sysDataDictionaryDetail.POST("createSysDataDictionaryDetail", system2.CreateSysDataDictionaryDetail)   // 新建SysDataDictionaryDetail
		sysDataDictionaryDetail.DELETE("deleteSysDataDictionaryDetail", system2.DeleteSysDataDictionaryDetail) // 新建SysDataDictionaryDetail
		sysDataDictionaryDetail.PUT("updateSysDataDictionaryDetail", system2.UpdateSysDataDictionaryDetail)    // 新建SysDataDictionaryDetail
		sysDataDictionaryDetail.GET("findSysDataDictionaryDetail", system2.FindSysDataDictionaryDetail)        // 新建SysDataDictionaryDetail
		sysDataDictionaryDetail.GET("getSysDataDictionaryDetailList", system2.GetSysDataDictionaryDetailList)  // 新建SysDataDictionaryDetail
	}
	Menu := System.Group("Menu")
	{
		Menu.GET("getMenu", (&system.SysAuthorityMenu{}).GetMenu)
	}

}
