package system

import (
	"ding/global"
	"ding/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SysMenu struct {
	SysBaseMenu
	MenuId      string    `json:"menuId" gorm:"comment:菜单ID"`
	AuthorityId uint      `json:"-" gorm:"comment:角色ID"`
	Children    []SysMenu `json:"children" gorm:"-"`
	//Parameters  []SysBaseMenuParameter `json:"parameters" gorm:"foreignKey:SysBaseMenuID;references:MenuId"`
	Btns map[string]uint `json:"btns" gorm:"-"`
}

type SysAuthorityMenu struct {
	MenuId      string `json:"menuId" gorm:"comment:菜单ID;column:sys_base_menu_id"`
	AuthorityId string `json:"-" gorm:"comment:角色ID;column:sys_authority_authority_id"`
}

func (s SysAuthorityMenu) TableName() string {
	return "sys_authority_menus"
}
func (a *SysAuthorityMenu) GetMenu(c *gin.Context) {
	//menus, err := (&SysBaseMenu{}).GetMenuTree(utils.GetUserAuthorityId(c))
	AuthorityId, _ := c.Get(global.CtxUserAuthorityIDKey)
	userID, _ := c.Get(global.CtxUserIDKey)
	fmt.Println(userID)
	menus, err := (&SysBaseMenu{}).GetMenuTree(AuthorityId.(uint))
	if err != nil {
		zap.L().Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
	}
	if menus == nil {
		menus = []SysMenu{}
	}
	response.OkWithDetailed(SysMenusResponse{Menus: menus}, "获取成功", c)
}
