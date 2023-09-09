package system

import (
	"ding/global"
	paramSystem "ding/model/params/system"
	"errors"
	"gorm.io/gorm"
)

type SysDataDictionary struct {
	gorm.Model
	Name                     string                    `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"`   // 字典名（中）
	Type                     string                    `json:"type" form:"type" gorm:"column:type;comment:字典名（英）"`   // 字典名（英）
	Status                   *bool                     `json:"status" form:"status" gorm:"column:status;comment:状态"` // 状态
	Desc                     string                    `json:"desc" form:"desc" gorm:"column:desc;comment:描述"`       // 描述
	SysDataDictionaryDetails []SysDataDictionaryDetail `json:"sysDataDictionaryDetails" form:"sysDataDictionaryDetails"`
}

func (SysDataDictionary) TableName() string {
	return "sys_data_dictionaries"
}

func (d *SysDataDictionary) CreateSysDataDictionary(sysDataDictionary SysDataDictionary) (err error) {
	if (!errors.Is(global.GLOAB_DB.First(SysDataDictionary{}, "type = ?", sysDataDictionary.Type).Error, gorm.ErrRecordNotFound)) {
		return errors.New("存在相同的type，不允许创建")
	}
	err = global.GLOAB_DB.Create(&sysDataDictionary).Error
	return err
}
func (d *SysDataDictionary) DeleteSysDataDictionary(sysDataDictionary SysDataDictionary) (err error) {
	err = global.GLOAB_DB.Where("id = ?", sysDataDictionary.ID).Preload("SysDataDictionaryDetails").First(&sysDataDictionary).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("请不要搞事")
	}
	if err != nil {
		return err
	}
	err = global.GLOAB_DB.Delete(&sysDataDictionary).Error
	if err != nil {
		return err
	}

	if sysDataDictionary.SysDataDictionaryDetails != nil {
		return global.GLOAB_DB.Where("sys_data_dictionary_id=?", sysDataDictionary.ID).Delete(sysDataDictionary.SysDataDictionaryDetails).Error
	}
	return
}
func (d *SysDataDictionary) UpdateSysDataDictionary(sysDataDictionary *SysDataDictionary) (err error) {
	var dict SysDataDictionary
	sysDataDictionaryMap := map[string]interface{}{
		"Name":   sysDataDictionary.Name,
		"Type":   sysDataDictionary.Type,
		"Status": sysDataDictionary.Status,
		"Desc":   sysDataDictionary.Desc,
	}
	db := global.GLOAB_DB.Where("id = ?", sysDataDictionary.ID).First(&dict)
	if dict.Type != sysDataDictionary.Type {
		if !errors.Is(global.GLOAB_DB.First(&SysDataDictionary{}, "type = ?", sysDataDictionary.Type).Error, gorm.ErrRecordNotFound) {
			return errors.New("存在相同的type，不允许创建")
		}
	}
	err = db.Updates(sysDataDictionaryMap).Error
	return err
}
func (d *SysDataDictionary) GetSysDataDictionary(Type string, Id uint, status *bool) (sysDataDictionary SysDataDictionary, err error) {
	var flag = false
	if status == nil {
		flag = true
	} else {
		flag = *status
	}
	err = global.GLOAB_DB.Where("(type = ? OR id = ?) and status = ?", Type, Id, flag).Preload("SysDataDictionaryDetails", "status = ?", true).First(&sysDataDictionary).Error
	return
}
func (d *SysDataDictionary) GetSysDataDictionaryInfoList(info paramSystem.SysDataDictionarySearch) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GLOAB_DB.Model(&SysDataDictionary{})
	var sysDataDictionarys []SysDataDictionary
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Name != "" {
		db = db.Where("`name` LIKE ?", "%"+info.Name+"%")
	}
	if info.Type != "" {
		db = db.Where("`type` LIKE ?", "%"+info.Type+"%")
	}
	if info.Status != nil {
		db = db.Where("`status` = ?", info.Status)
	}
	if info.Desc != "" {
		db = db.Where("`desc` LIKE ?", "%"+info.Desc+"%")
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Find(&sysDataDictionarys).Error
	return sysDataDictionarys, total, err
}
