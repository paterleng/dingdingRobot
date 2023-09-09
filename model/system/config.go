package system

import (
	"ding/global"
	"gorm.io/gorm"
)

type Config struct {
	gorm.Model
	Key   string `json:"key"`
	Value int    `json:"value"`
	Desc  string `json:"desc"`
}

func (c *Config) GetConfig(key string) (*Config, error) {
	err := global.GLOAB_DB.First(c, "key = ?", key).Error
	return c, err
}
