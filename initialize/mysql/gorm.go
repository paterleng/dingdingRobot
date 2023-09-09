package mysql

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
)

func RegisterTables(db *gorm.DB) (err error) {
	//err = db.AutoMigrate(&dingding2.UserDept{})
	//err = db.AutoMigrate(&dingding2.DingUser{})
	//err = db.AutoMigrate(&dingding2.DingRobot{})
	//err = db.AutoMigrate(&dingding2.DingDept{})
	//err = db.AutoMigrate(&dingding2.DingAttendGroup{})
	//err = db.AutoMigrate(&dingding2.RestTime{})
	//err = db.AutoMigrate(&dingding2.Task{})
	//err = db.AutoMigrate(&common.MsgText{})
	//err = db.AutoMigrate(&common.MsgLink{})
	//err = db.AutoMigrate(&common.MsgMarkDown{})
	//err = db.AutoMigrate(&common.MarkDown{})
	//err = db.AutoMigrate(&common.At{})
	//err = db.AutoMigrate(&common.Text{})
	//err = db.AutoMigrate(&common.AtMobile{})
	//err = db.AutoMigrate(&common.AtUserId{})
	//err = db.AutoMigrate(&system.Config{})
	//err = db.AutoMigrate(&system.SysDataDictionary{})
	//err = db.AutoMigrate(&system.SysDataDictionaryDetail{})
	//err = db.AutoMigrate(&system.SysBaseMenu{})
	//err = db.AutoMigrate(&system.SysAuthority{})
	//err = db.AutoMigrate(&system.SysBaseMenuBtn{})
	//err = db.AutoMigrate(&system.SysBaseMenuParameter{})
	//err = db.AutoMigrate(&system.SysAuthorityBtn{})
	//err = db.AutoMigrate(&dingding2.SubscriptionRelationship{})

	//err = db.AutoMigrate(
	//	dingding2.DingUser{},
	//	dingding2.DingRobot{},
	//	dingding2.DingDept{},
	//	dingding2.DingAttendGroup{},
	//	dingding2.Task{},
	//	common.MsgText{},
	//	common.MsgLink{},
	//	common.MsgMarkDown{},
	//	common.MarkDown{},
	//	common.At{},
	//	common.Text{},
	//	common.AtMobile{},
	//	system.Config{},
	//	system.SysDataDictionary{},
	//	system.SysDataDictionaryDetail{},
	//)
	if err != nil {
		zap.L().Error("register table failed", zap.Error(err))
		os.Exit(0)
	}
	zap.L().Info("register table success")
	return err
}
