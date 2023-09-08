package model

import (
	"ding/global"
	"gorm.io/gorm"
)

type Config struct {
	gorm.Model
	Key         string `json:"mysql:key"`
	Value       string
	Description string
}

const ConfigTableName = "config"

func (co *Config) TableName() string {
	return ConfigTableName
}

func (co *Config) GetConfigInfoByKey(key string) (err error) {
	return global.GLOAB_DB.Table(co.TableName()).Where("`key` = ? And deleted_at=0", key).First(&co).Error
}
func GetConfigValue(key string) (value string) {
	// 传进来一个key 可以把config表中对应value返回
	var configStruct Config
	err := global.GLOAB_DB.Table(configStruct.TableName()).Select("value").Where("`key` = ? AND deleted_at=0", key).First(&configStruct).Error
	if err != nil {
		return ""
	}
	return configStruct.Value
}
