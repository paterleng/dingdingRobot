package system

import (
	"ding/global"
	request "ding/model/params/system"
	"gorm.io/gorm"
)

// 如果含有time.Time 请自行import time包
type SysDataDictionaryDetail struct {
	gorm.Model
	Label               string `json:"label" form:"label" gorm:"column:label;comment:展示值"`                                               // 展示值
	Value               int    `json:"value" form:"value" gorm:"column:value;comment:字典值"`                                               // 字典值
	Status              *bool  `json:"status" form:"status" gorm:"column:status;comment:启用状态"`                                           // 启用状态
	Sort                int    `json:"sort" form:"sort" gorm:"column:sort;comment:排序标记"`                                                 // 排序标记
	SysDataDictionaryID int    `json:"sysDataDictionaryID" form:"sysDataDictionaryID" gorm:"column:sys_data_dictionary_id;comment:关联标记"` // 关联标记
}

func (SysDataDictionaryDetail) TableName() string {
	return "sys_data_dictionary_details"
}
func (dictionaryDetailService *SysDataDictionaryDetail) CreateSysDataDictionaryDetail(sysDataDictionaryDetail SysDataDictionaryDetail) (err error) {
	err = global.GLOAB_DB.Create(&sysDataDictionaryDetail).Error
	return err
}

//@author: RZZY
//@function: DeleteSysDataDictionaryDetail
//@description: 删除字典详情数据
//@param: sysDataDictionaryDetail model.SysDataDictionaryDetail
//@return: err error

func (dictionaryDetailService *SysDataDictionaryDetail) DeleteSysDataDictionaryDetail(sysDataDictionaryDetail SysDataDictionaryDetail) (err error) {
	err = global.GLOAB_DB.Delete(&sysDataDictionaryDetail).Error
	return err
}

//@author: RZZY
//@function: UpdateSysDataDictionaryDetail
//@description: 更新字典详情数据
//@param: sysDataDictionaryDetail *model.SysDataDictionaryDetail
//@return: err error

func (dictionaryDetailService *SysDataDictionaryDetail) UpdateSysDataDictionaryDetail(sysDataDictionaryDetail *SysDataDictionaryDetail) (err error) {
	err = global.GLOAB_DB.Updates(sysDataDictionaryDetail).Error
	return err
}

//@author: RZZY
//@function: GetSysDataDictionaryDetail
//@description: 根据id获取字典详情单条数据
//@param: id uint
//@return: sysDataDictionaryDetail SysDataDictionaryDetail, err error

func (dictionaryDetailService *SysDataDictionaryDetail) GetSysDataDictionaryDetail(sysDataDictionaryID int) (sysDataDictionaryDetail []SysDataDictionaryDetail, err error) {
	err = global.GLOAB_DB.Where("sys_data_dictionary_id = ?", sysDataDictionaryID).Find(&sysDataDictionaryDetail).Error
	return
}

//@author: RZZY
//@function: GetSysDataDictionaryDetailInfoList
//@description: 分页获取字典详情列表
//@param: info request.SysDataDictionaryDetailSearch
//@return: list interface{}, total int64, err error

func (dictionaryDetailService *SysDataDictionaryDetail) GetSysDataDictionaryDetailInfoList(info request.SysDataDictionaryDetailSearch) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GLOAB_DB.Model(&SysDataDictionaryDetail{})
	var sysDataDictionaryDetails []SysDataDictionaryDetail
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Label != "" {
		db = db.Where("label LIKE ?", "%"+info.Label+"%")
	}
	if info.Value != 0 {
		db = db.Where("value = ?", info.Value)
	}
	if info.Status != nil {
		db = db.Where("status = ?", info.Status)
	}
	if info.SysDataDictionaryID != 0 {
		db = db.Where("sys_data_dictionary_id = ?", info.SysDataDictionaryID)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Order("sort").Find(&sysDataDictionaryDetails).Error
	return sysDataDictionaryDetails, total, err
}
