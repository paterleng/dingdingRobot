package request

import (
	"ding/model/common/request"
)

type SysDataDictionaryDetailSearch struct {
	SysDataDictionaryID uint   `json:"sys_data_dictionary_id"`
	Label               string `json:"label" form:"label" gorm:"column:label;comment:展示值"`     // 展示值
	Value               int    `json:"value" form:"value" gorm:"column:value;comment:字典值"`     // 字典值
	Status              *bool  `json:"status" form:"status" gorm:"column:status;comment:启用状态"` // 启用状态
	Sort                int    `json:"sort" form:"sort" gorm:"column:sort;comment:排序标记"`       // 排序标记
	request.PageInfo
}
